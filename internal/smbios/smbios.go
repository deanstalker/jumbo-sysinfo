package smbios

import (
	"fmt"

	dsmbios "github.com/digitalocean/go-smbios/smbios"
	tsmbios "github.com/talos-systems/go-smbios/smbios"
)

type SMBIOS struct {
	tsmbios.SMBIOS

	MemoryDeviceStructure []tsmbios.MemoryDeviceStructure
}

type MemoryDeviceStructure struct {
	tsmbios.MemoryDeviceStructure
}

// MemoryDevice returns a `MemoryDeviceStructure`.
func (s *SMBIOS) MemoryDevice() []tsmbios.MemoryDeviceStructure {
	return s.MemoryDeviceStructure
}

func New() (*SMBIOS, error) {
	rc, ep, err := dsmbios.Stream()
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	//nolint: errcheck
	defer rc.Close()

	s := &SMBIOS{}

	s.Version.Major, s.Version.Minor, s.Version.Revision = ep.Version()

	d := dsmbios.NewDecoder(rc)

	ss, err := d.Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode structures: %w", err)
	}

	s.Structures = ss

	for _, structure := range s.Structures {
		switch structure.Header.Type {
		case 0:
			s.BIOSInformationStructure = tsmbios.BIOSInformationStructure{Structure: structure}
		case 1:
			s.SystemInformationStructure = tsmbios.SystemInformationStructure{Structure: structure}
		case 2:
			s.BaseboardInformationStructure = tsmbios.BaseboardInformationStructure{Structure: structure}
		case 3:
			s.SystemEnclosureStructure = tsmbios.SystemEnclosureStructure{Structure: structure}
		case 4:
			s.ProcessorInformationStructure = tsmbios.ProcessorInformationStructure{Structure: structure}
		case 5:
			// Obsolete.
		case 6:
			// Obsolete.
		case 7:
			s.CacheInformationStructure = tsmbios.CacheInformationStructure{Structure: structure}
		case 8:
			s.PortConnectorInformationStructure = tsmbios.PortConnectorInformationStructure{Structure: structure}
		case 9:
			s.SystemSlotsStructure = tsmbios.SystemSlotsStructure{Structure: structure}
		case 10:
			// Obsolete.
		case 11:
			s.OEMStringsStructure = tsmbios.OEMStringsStructure{Structure: structure}
		case 12:
			s.SystemConfigurationOptionsStructure = tsmbios.SystemConfigurationOptionsStructure{Structure: structure}
		case 13:
			s.BIOSLanguageInformationStructure = tsmbios.BIOSLanguageInformationStructure{Structure: structure}
		case 14:
			s.GroupAssociationsStructure = tsmbios.GroupAssociationsStructure{Structure: structure}
		case 15:
			// Unimplemented.
		case 16:
			s.PhysicalMemoryArrayStructure = tsmbios.PhysicalMemoryArrayStructure{Structure: structure}
		case 17:
			s.MemoryDeviceStructure = append(s.MemoryDeviceStructure, tsmbios.MemoryDeviceStructure{Structure: structure})
		}
	}

	return s, nil
}
