package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"jumbo-sysinfo/internal/smbios"
	"jumbo-sysinfo/internal/utils/powershell"
)

type SMI struct {
	SystemInformation    SystemInformation    `json:"system_information"`
	BaseBoardInformation BaseBoardInformation `json:"base_board_information"`
	ProcessorInformation ProcessorInformation `json:"processor_information"`
	MemoryDevices        []MemoryDevice       `json:"memory_devices"`
	Disks                []Disk               `json:"disks"`
}

type Disk struct {
	Path         string `json:"path"`
	FriendlyName string `json:"friendlyName"`
	BusType      string `json:"busType"`
	Size         string `json:"size"`
}

type SystemInformation struct {
	Family       string `json:"family"`
	Manufacturer string `json:"manufacturer"`
	ProductName  string `json:"productName"`
	SerialNumber string `json:"serialNumber"`
	SKU          string `json:"sku"`
	Version      string `json:"version"`
}

type BaseBoardInformation struct {
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	Version      string `json:"version"`
	SerialNumber string `json:"serialNumber"`
	AssetTag     string `json:"assetTag"`
}

type ProcessorInformation struct {
	ProcessorVersion  string `json:"processorVersion"`
	SocketDesignation string `json:"socketDesignation"`
	SerialNumber      string `json:"serialNumber"`
	PartNumber        string `json:"partNumber"`
	CurrentSpeed      string `json:"currentSpeed"`
	CoreCount         string `json:"coreCount"`
	ThreadCount       string `json:"threadCount"`
}

type MemoryDevice struct {
	Manufacturer          string `json:"manufacturer"`
	SerialNumber          string `json:"serialNumber"`
	BankLocator           string `json:"bankLocator"`
	ConfiguredMemorySpeed string `json:"configuredMemorySpeed"`
	ConfiguredVoltage     string `json:"configuredVoltage"`
	MemoryType            string `json:"memoryType"`
	DataWidth             string `json:"dataWidth"`
	FormFactor            string `json:"formFactor"`
	Size                  string `json:"size"`
	PartNumber            string `json:"partNumber"`
}

type GetDisk struct {
	FriendlyName string `json:"FriendlyName"`
	DiskNumber   int    `json:"DiskNumber"`
	BusType      string `json:"BusType"`
	Size         int64  `json:"Size"`
}

func main() {
	OSType := runtime.GOOS
	switch OSType {
	case "darwin":
		panic("OSX is not supported")
	}

	// Find SMBIOS data in operating system-specific location.
	sm, err := smbios.New()
	if err != nil {
		panic(err.Error())
	}

	var smi SMI
	smi.SystemInformation = SystemInformation{
		Family:       sm.SystemInformation().Family(),
		Manufacturer: sm.SystemInformation().Manufacturer(),
		ProductName:  sm.SystemInformation().ProductName(),
		SerialNumber: sm.SystemInformation().SerialNumber(),
		SKU:          sm.SystemInformation().SKUNumber(),
		Version:      sm.SystemInformation().Version(),
	}

	smi.BaseBoardInformation = BaseBoardInformation{
		Manufacturer: sm.BaseboardInformation().Manufacturer(),
		Product:      sm.BaseboardInformation().Product(),
		Version:      sm.BaseboardInformation().Version(),
		SerialNumber: sm.BaseboardInformation().SerialNumber(),
		AssetTag:     sm.BaseboardInformation().AssetTag(),
	}

	smi.ProcessorInformation = ProcessorInformation{
		SocketDesignation: sm.ProcessorInformation().SocketDesignation(),
		SerialNumber:      sm.ProcessorInformation().SerialNumber(),
		PartNumber:        sm.ProcessorInformation().PartNumber(),
		ProcessorVersion:  sm.ProcessorInformation().ProcessorVersion(),
		CurrentSpeed:      fmt.Sprintf("%d Mhz", binary.LittleEndian.Uint16(sm.ProcessorInformation().Formatted[18:20])),
		CoreCount:         fmt.Sprintf("%d", sm.ProcessorInformation().Formatted[31]),
		ThreadCount:       fmt.Sprintf("%d", sm.ProcessorInformation().Formatted[33]),
	}

	memDevices := sm.MemoryDevice()
	for _, memDevice := range memDevices {
		smi.MemoryDevices = append(smi.MemoryDevices, MemoryDevice{
			Manufacturer:          memDevice.Manufacturer(),
			SerialNumber:          memDevice.SerialNumber(),
			BankLocator:           memDevice.BankLocator(),
			ConfiguredMemorySpeed: memDevice.ConfiguredMemorySpeed().String(),
			ConfiguredVoltage:     memDevice.ConfiguredVoltage().String(),
			MemoryType:            memDevice.MemoryType().String(),
			DataWidth:             memDevice.DataWidth().String(),
			FormFactor:            memDevice.FormFactor().String(),
			Size:                  fmt.Sprintf("%d GB", binary.LittleEndian.Uint16(memDevice.Formatted[8:10])/1000),
			PartNumber:            memDevice.PartNumber(),
		})
	}

	if OSType == "windows" {
		ps := powershell.New()

		args := []string{"Get-Disk", "|", "ConvertTo-Json"}
		stdOut, _, err := ps.Execute(args...)
		if err != nil {
			panic(err.Error())
		}

		var gd GetDisk
		if err := json.Unmarshal([]byte(stdOut), &gd); err != nil {
			panic(err.Error())
		}

		disk := Disk{
			Path:         fmt.Sprintf("//%d", gd.DiskNumber),
			FriendlyName: gd.FriendlyName,
			BusType:      gd.BusType,
			Size:         fmt.Sprintf("%d GB", gd.Size/1024/1024/1000),
		}
		smi.Disks = append(smi.Disks, disk)
	}

	if OSType == "linux" {
		lsblkPath, err := exec.LookPath("lsblk")
		if err != nil {
			panic(err.Error())
		}
		cmd := exec.Command(lsblkPath, []string{"-d", "-e", "7", "-o", "NAME,MODEL,VENDOR,TRAN,SIZE"}...)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			panic(err.Error())
		}
		stdOut := stdout.String()

		rows := strings.Split(stdOut, "\n")
		body := rows[1:]
		for _, row := range body {
			data := strings.Fields(row)
			if len(data) == 0 {
				continue
			}

			disk := Disk{
				Path:         data[0],
				FriendlyName: fmt.Sprintf("%s %s", data[2], data[1]),
				BusType:      data[3],
				Size:         data[4],
			}
			smi.Disks = append(smi.Disks, disk)
		}
	}

	data, err := json.Marshal(smi)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", data)
}
