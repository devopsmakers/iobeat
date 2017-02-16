package beater

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/devopsmakers/iobeat/config"
)

type Iobeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

type DiskStats struct {
	Major             int
	Minor             int
	Name              string
	ReadRequests      uint64 // Total number of reads completed successfully.
	ReadMerged        uint64 // Adjacent read requests merged in a single req.
	ReadSectors       uint64 // Total number of sectors read successfully.
	MsecRead          uint64 // Total number of ms spent by all reads.
	WriteRequests     uint64 // total number of writes completed successfully.
	WriteMerged       uint64 // Adjacent write requests merged in a single req.
	WriteSectors      uint64 // total number of sectors written successfully.
	MsecWrite         uint64 // Total number of ms spent by all writes.
	IosInProgress     uint64 // Number of actual I/O requests currently in flight.
	MsecTotal         uint64 // Amount of time during which ios_in_progress >= 1.
	MsecWeightedTotal uint64 // Measure of recent I/O completion time and backlog.
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Iobeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Iobeat) Run(b *beat.Beat) error {
	logp.Info("iobeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		events, err := bt.CollectIOStats()
		if err != nil {
			log.Fatal(err)
		}

		bt.client.PublishEvents(events)

		logp.Info("Events sent")

	}
}

func (bt *Iobeat) CollectIOStats() ([]common.MapStr, error) {

	events := make([]common.MapStr, 0)

	if diskstats, err := os.Open("/proc/diskstats"); err == nil {
		defer diskstats.Close()
		diskscan := bufio.NewScanner(diskstats)

		for diskscan.Scan() {
			if err = diskscan.Err(); err != nil {
				return nil, err
			}

			fields := strings.Fields(diskscan.Text())

			// Get rid of devices with no stats
			if fields[3] == "0" {
				continue
			}

			// Ensure the output is as expected
			size := len(fields)
			if size != 14 {
				return nil, fmt.Errorf("OS Kernel version too low!")
			}

			if bt.config.Disks != nil {
				for _, disk := range *bt.config.Disks {
					if fields[2] == disk {
						// For specified disks
						events = append(events, MakeEvent(fields))
					}
				}
			} else {
				// Default: all disks
				events = append(events, MakeEvent(fields))
			}

		}

	} else {
		return nil, err
	}

	return events, nil
}

func MakeEvent(fields []string) common.MapStr {
	size := len(fields)

	diskioevent := DiskStats{}

	for i := 0; i < size; i++ {
		diskioevent.Major, _ = strconv.Atoi(fields[0])
		diskioevent.Minor, _ = strconv.Atoi(fields[1])
		diskioevent.Name = fields[2]
		diskioevent.ReadRequests, _ = strconv.ParseUint(fields[3], 10, 64)
		diskioevent.ReadMerged, _ = strconv.ParseUint(fields[4], 10, 64)
		diskioevent.ReadSectors, _ = strconv.ParseUint(fields[5], 10, 64)
		diskioevent.MsecRead, _ = strconv.ParseUint(fields[6], 10, 64)
		diskioevent.WriteRequests, _ = strconv.ParseUint(fields[7], 10, 64)
		diskioevent.WriteMerged, _ = strconv.ParseUint(fields[8], 10, 64)
		diskioevent.WriteSectors, _ = strconv.ParseUint(fields[9], 10, 64)
		diskioevent.MsecWrite, _ = strconv.ParseUint(fields[10], 10, 64)
		diskioevent.IosInProgress, _ = strconv.ParseUint(fields[11], 10, 64)
		diskioevent.MsecTotal, _ = strconv.ParseUint(fields[12], 10, 64)
		diskioevent.MsecWeightedTotal, _ = strconv.ParseUint(fields[13], 10, 64)
	}

	event := common.MapStr{
		"@timestamp": common.Time(time.Now()),
		"type":       "diskstats",
		"stats": common.MapStr{
			"major":               diskioevent.Major,
			"minor":               diskioevent.Minor,
			"name":                diskioevent.Name,
			"read_requests":       diskioevent.ReadRequests,
			"read_merged":         diskioevent.ReadMerged,
			"read_sectors":        diskioevent.ReadSectors,
			"msec_read":           diskioevent.MsecRead,
			"write_requests":      diskioevent.WriteRequests,
			"write_merged":        diskioevent.WriteMerged,
			"write_sectors":       diskioevent.WriteSectors,
			"msec_write":          diskioevent.MsecWrite,
			"ios_in_progress":     diskioevent.IosInProgress,
			"msec_total":          diskioevent.MsecTotal,
			"msec_weighted_total": diskioevent.MsecWeightedTotal,
		},
	}

	return event
}

func (bt *Iobeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
