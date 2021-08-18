package svd

import (
	"encoding/xml"
)

type AccessType string

const (
	AccessReadOnly      AccessType = "read-only"
	AccessWriteOnly     AccessType = "write-only"
	AccessReadWrite     AccessType = "read-write"
	AccessWriteOnce     AccessType = "writeOnce"
	AccessReadWriteOnce AccessType = "read-writeOnce"
)

type ProtectionType string

const (
	ProtectionSecure     ProtectionType = "s"
	ProtectionNonSecure  ProtectionType = "n"
	ProtectionPrivileged ProtectionType = "p"
)

type Peripherals struct {
	// Define the sequence of peripherals.
	Peripheral []Peripheral `xml:"peripheral"`
}

type Device struct {
	XMLName xml.Name `xml:"device"`

	// Specify the underlying XML schema to which the CMSIS-SVD schema is compliant.
	Xs string `xml:"xmlns:xs,attr"`
	// Specify the file path and file name of the CMSIS-SVD Schema.
	NoNamespaceSchemaLocation string `xml:"xs:noNamespaceSchemaLocation,attr"`
	// Specify the compliant CMSIS-SVD schema version.
	SchemaVersion string `xml:"schemaVersion,attr"`

	// Specify the vendor of the device using the full name.
	Vendor string `xml:"vendor,omitempty"`

	// Specify the vendor abbreviation without spaces or special characters.
	// This information is used to define the directory.
	VendorID string `xml:"vendorID,omitempty"`

	// The string identifies the device or device series.
	// Device names are required to be unique.
	Name string `xml:"name"`

	// Specify the name of the device series.
	Series string `xml:"series,omitempty"`

	// Define the version of the SVD file.
	// Silicon vendors maintain the description throughout the life-cycle of the device
	// and ensure that all updated and released copies have a unique version string.
	// Higher numbers indicate a more recent version.
	Version string `xml:"version"`

	// Describe the main features of the device
	// (for example CPU, clock frequency, peripheral overview).
	Description string `xml:"description"`

	// The text will be copied into the header section of the generated device header
	// file and shall contain the legal disclaimer.
	// New lines can be inserted by using \n.
	// This section is mandatory if the SVD file is used for generating the device header file.
	LicenseText string `xml:"licenseText,omitempty"`

	// Describe the processor included in the device.
	Cpu Cpu `xml:"cpu,omitempty"`

	// Specify the file name (without extension) of the device-specific system
	// include file (system_<device>.h; See CMSIS-Core description).
	// The header file generator customizes the include statement referencing the
	// CMSIS system file within the CMSIS device header file.
	// By default, the filename is system_device-name.h.
	// In cases where a device series shares a single system header file,
	// the name of the series shall be used instead of the individual device name.
	HeaderSystemFilename string `xml:"headerSystemFilename,omitempty"`

	// This string is prepended to all type definition names generated in the
	// CMSIS-Core device header file.
	// This is used if the vendor's software requires vendor-specific types in
	// order to avoid name clashes with other definied types.
	HeaderDefinitionsPrefix string `xml:"headerDefinitionsPrefix,omitempty"`

	// Define the number of data bits uniquely selected by each address.
	// The value for Cortex-M-based devices is 8 (byte-addressable).
	AddressUnitBits uint `xml:"addressUnitBits"`

	// Define the number of data bit-width of the maximum single data transfer
	// supported by the bus infrastructure.
	// This information is relevant for debuggers when accessing registers,
	// because it might be required to issue multiple accesses for resources of
	// a bigger size.
	// The expected value for Cortex-M-based devices is 32.
	Width uint `xml:"width"`

	// Default bit-width of any register contained in the device.
	Size uint `xml:"size,omitempty"`

	// Default access rights for all registers.
	Access AccessType `xml:"access,omitempty"`

	// Default access protection for all registers.
	Protection ProtectionType `xml:"protection,omitempty"`

	// Default value for all registers at RESET.
	ResetValue string `xml:"resetValue,omitempty"`

	// Define which register bits have a defined reset value.
	ResetMask string `xml:"resetMask,omitempty"`

	// Group to define peripherals.
	Peripherals Peripherals `xml:"peripherals"`

	// The content and format of this section is unspecified.
	// Silicon vendors may choose to provide additional information.
	// By default, this section is ignored when constructing CMSIS files.
	// It is up to the silicon vendor to specify a schema for this section.
	VendorExtensions string `xml:"vendorExtensions,omitempty"`
}

func NewDevice(name string) *Device {
	dev := Device{
		SchemaVersion:             "1.3.6",
		Xs:                        "http://www.w3.org/2001/XMLSchema-instance",
		NoNamespaceSchemaLocation: "CMSIS-SVD.xsd",
		Name:                      name,
		AddressUnitBits:           8,
		Width:                     32,
		Size:                      32,
		Access:                    AccessReadWrite,
		ResetValue:                "0x00000000",
		ResetMask:                 "0xFFFFFFFF",
	}
	return &dev
}

func (dev Device) SVD() (svd []byte, err error) {
	svd, err = xml.MarshalIndent(dev, "", "  ")
	return append([]byte(xml.Header), svd...), err
}
