package main

import (
	"flag"
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/hashicorp/logutils"
	"github.com/hbouvier/httpclient"
	"log"
	"os"
	"time"
)

type Logstash struct {
	Tag      string `json:"tag"`
	Hostname string `json:"hostname"`
	Metric   Metric `json:"metric"`
	Message  string `json:"message"`
}

type Metric struct {
	Cpu    CPU             `json:"cpu"`
	Memory Memory          `json:"memory"`
	Disks  map[string]Disk `json:"disks"`
}

type CPU struct {
	Total int `json:"total"`
}

type Memory struct {
	Total_mb int `json:"total_mb"`
	Free_mb  int `json:"free_mb"`
	Avail_mb int `json:"avail_mb"`
	Free     int `json:"free"`
}

type Disk struct {
	Size_gb int `json:"size_gb"`
	Used_gb int `json:"used_gb"`
	Free    int `json:"free"`
}

type CPUSample struct {
	idle_ticks  uint64
	total_ticks uint64
}

func main() {
	logstashPtr := flag.String("logstash", "http://logstash:31311", "Logstash http endpoint (default: http://logstash:31311)")
	intervalPtr := flag.Int("interval", 60, "Publish the resources statistic at this interval in seconds (default: 60)")
	hostnamePtr := flag.String("hostname", "unknown", "Name of the host")
	levelPtr := flag.String("level", "INFO", "Logging level (default: INFO)")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel(*levelPtr),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	if *hostnamePtr == "unknown" {
		fmt.Printf("missing option '-hostname'")
		usage(flag.Args())
	}

	volumes := make([]string, 0)
	switch len(flag.Args()) {
	case 0:
		volumes = append(volumes, "/")
	default:
		for i := range flag.Args() {
			volumes = append(volumes, flag.Args()[i])
		}
	}
	log.Printf("[DEBUG] logstash=%s interval=%d", *logstashPtr, *intervalPtr)
	monitor(*logstashPtr, *intervalPtr, *hostnamePtr, volumes)
}

func usage(args []string) {
	fmt.Printf("USAGE: monitor -hostname MyServer -logstash http://logstash:31311 -interval 60 -level INFO {/ /mount/disk ...}\n")
	os.Exit(2)
}

func monitor(logstashURL string, interval int, hostname string, volumes []string) {
	client := httpclient.New(logstashURL, nil, map[string]string{"Content-Type": "application/json; charset=UTF-8"})
	last_sample := sample_cpu()
	time.Sleep(time.Duration(1000) * time.Millisecond)
	for {
		new_sample := sample_cpu()
		cpu := cpu(last_sample, new_sample)
		last_sample = new_sample
		mem := memory()
		message := fmt.Sprintf("host: %s, CPU usage: %d%%", hostname, cpu.Total)
		disks := map[string]Disk{}
		for i := range volumes {
			disks[volumes[i]] = disk(volumes[i])
			message = message + fmt.Sprintf(", disk %s %d%% free", volumes[i], disks[volumes[i]].Free)
		}
		log.Printf("[DEBUG] %s\n", message)

		var response string
		payload := Logstash{Hostname: hostname, Tag: "cluster_watch", Message: message, Metric: Metric{Cpu: cpu, Memory: mem, Disks: disks}}
		err := client.Post("/", payload, &response)
		if err != nil {
			log.Printf("[ERROR] POST %s >>> %v", logstashURL, err)
			os.Exit(-1)
		}
		log.Printf("[INFO] %s => %s\n", message, response)
		time.Sleep(time.Duration(interval*1000) * time.Millisecond)
	}
}

func sample_cpu() CPUSample {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.Fatal("Unable to read /proc/stat >>> %v", err)
		os.Exit(-1)
	}

	idle := stat.CPUStatAll.Idle
	total := stat.CPUStatAll.User +
		stat.CPUStatAll.Nice +
		stat.CPUStatAll.System +
		stat.CPUStatAll.Idle +
		stat.CPUStatAll.IOWait +
		stat.CPUStatAll.IRQ +
		stat.CPUStatAll.SoftIRQ +
		stat.CPUStatAll.Steal +
		stat.CPUStatAll.Guest +
		stat.CPUStatAll.GuestNice
	return CPUSample{idle_ticks: idle, total_ticks: total}
}

func cpu(last CPUSample, current CPUSample) CPU {
	totalTicks := float64(current.total_ticks - last.total_ticks)
	idleTicks := float64(current.idle_ticks - last.idle_ticks)
	return CPU{Total: int(100.0 * (totalTicks - idleTicks) / totalTicks)}
}

func memory() Memory {
	mem, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		log.Fatal("Unable to read /proc/meminfo >>> %v", err)
		os.Exit(-1)
	}
	mb := uint64(1024)
	return Memory{Total_mb: int(mem.MemTotal / mb), Free_mb: int(mem.MemFree / mb), Avail_mb: int(mem.MemAvailable / mb), Free: int(mem.MemFree * uint64(100) / mem.MemTotal)}
}

// / and /var/lib/automat/couchdb
func disk(path string) Disk {
	disk, err := linuxproc.ReadDisk(path)
	if err != nil {
		log.Fatal("Unable to read %s >>> %v", path, err)
		os.Exit(-1)
	}
	gb := uint64(1024 * 1024 * 1024)
	return Disk{Size_gb: int(disk.All / gb), Used_gb: int(disk.Used / gb), Free: int(((disk.All - disk.Used) * 100) / disk.All)}
}
