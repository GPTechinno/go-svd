package svd

type CpuName string

const (
	// Arm Cortex-M0
	CpuNameCM0 CpuName = "CM0"
	// Arm Cortex-M0+
	CpuNameCM0p CpuName = "CM0+"
	// Arm Cortex-M1
	CpuNameCM1 CpuName = "CM1"
	// Arm Secure Core SC000
	CpuNameSC000 CpuName = "SC000"
	// Arm Cortex-M23
	CpuNameCM23 CpuName = "CM23"
	// Arm Cortex-M3
	CpuNameCM3 CpuName = "CM3"
	// Arm Cortex-M33
	CpuNameCM33 CpuName = "CM33"
	// Arm Cortex-M35P
	CpuNameCM35P CpuName = "CM35P"
	// Arm Cortex-M55
	CpuNameCM55 CpuName = "CM55"
	// Arm Secure Core SC300
	CpuNameSC300 CpuName = "SC300"
	// Arm Cortex-M4
	CpuNameCM4 CpuName = "CM4"
	// Arm Cortex-M7
	CpuNameCM7 CpuName = "CM7"
	// Arm Cortex-A5
	CpuNameCA5 CpuName = "CA5"
	// Arm Cortex-A7
	CpuNameCA7 CpuName = "CA7"
	// Arm Cortex-A8
	CpuNameCA8 CpuName = "CA8"
	// Arm Cortex-A9
	CpuNameCA9 CpuName = "CA9"
	// Arm Cortex-A15
	CpuNameCA15 CpuName = "CA15"
	// Arm Cortex-A17
	CpuNameCA17 CpuName = "CA17"
	// Arm Cortex-A53
	CpuNameCA53 CpuName = "CA53"
	// Arm Cortex-A57
	CpuNameCA57 CpuName = "CA57"
	// Arm Cortex-A72
	CpuNameCA72 CpuName = "CA72"
	// other processor architectures
	CpuNameother CpuName = "other"
)

type EndianType string

const (
	// little endian memory
	// (least significant byte gets allocated at the lowest address).
	EndianLittle EndianType = "little"
	// byte invariant big endian data organization
	// (most significant byte gets allocated at the lowest address).
	EndianBig EndianType = "big"
	// little and big endian are configurable for the device
	// and become active after the next reset.
	EndianSelectable EndianType = "selectable"
	// the endianness is neither little nor big endian.
	EndianOther EndianType = "other"
)

type RegionAccessType string

const (
	RegionAccessNonSecure      RegionAccessType = "n"
	RegionAccessSecureCallable RegionAccessType = "c"
)

type Region struct {
	// Specify whether the Secure Attribution Units are enabled.
	// Default value is true.
	Enabled bool `xml:"enabled,attr,omitempty"`

	// Identifiy the region with a name.
	Name string `xml:"name,attr,omitempty"`

	// Base address of the region.
	Base uint `xml:"base"`

	// Limit address of the region.
	Limit uint `xml:"limit"`

	// Define the acces type of a region.
	Access RegionAccessType `xml:"access"`
}

type SauRegionsConfigType struct {
	// Specify whether the Secure Attribution Units are enabled.
	Enabled bool `xml:"enabled,attr,omitempty"`

	// Set the protection mode for disabled regions.
	// When the complete SAU is disabled, the whole memory is treated
	// either "s"=secure or "n"=non-secure.
	// This value is inherited by the <region> element.
	ProtectionWhenDisabled ProtectionType `xml:"protectionWhenDisabled,attr,omitempty"`

	// Group to configure SAU regions.
	Region []Region `xml:"region,omitempty"`
}

type Cpu struct {
	Name CpuName `xml:"name"`

	// Define the HW revision of the processor.
	// The version format is rNpM (N,M = [0 - 99]).
	Revision string `xml:"revision"`

	// Define the endianness of the processor.
	Endian EndianType `xml:"endian"`

	// Indicate whether the processor is equipped with a
	// memory protection unit (MPU).
	MpuPresent bool `xml:"mpuPresent"`

	// Indicate whether the processor is equipped with a
	// hardware floating point unit (FPU).
	// Cortex-M4, Cortex-M7, Cortex-M33 and Cortex-M35P are
	// the only available Cortex-M processor with an optional FPU.
	FpuPresent bool `xml:"fpuPresent"`

	// Indicate whether the processor is equipped with a
	// double precision floating point unit.
	// This element is valid only when <fpuPresent> is set to true.
	// Currently, only Cortex-M7 processors can have a
	// double precision floating point unit.
	FpuDP bool `xml:"fpuDP,omitempty"`

	// Indicates whether the processor implements the optional
	// SIMD DSP extensions (DSP).
	// Cortex-M33 and Cortex-M35P are the only available Cortex-M
	// processor with an optional DSP extension.
	// For ARMv7M SIMD DSP extensions are a mandatory part of
	// Cortex-M4 and Cortex-M7.
	// This element is mandatory for Cortex-M33, Cortex-M35P
	// and future processors with optional SIMD DSP instruction set.
	DspPresent bool `xml:"dspPresent,omitempty"`

	// Indicate whether the processor has an instruction cache.
	// Note: only for Cortex-M7-based devices.
	IcachePresent bool `xml:"icachePresent,omitempty"`

	// Indicate whether the processor has a data cache.
	// Note: only for Cortex-M7-based devices.
	DcachePresent bool `xml:"dcachePresent,omitempty"`

	// Indicate whether the processor has an instruction
	// tightly coupled memory.
	// Note: only an option for Cortex-M7-based devices.
	ItcmPresent bool `xml:"itcmPresent,omitempty"`

	// Indicate whether the processor has a data tightly
	// coupled memory.
	// Note: only for Cortex-M7-based devices.
	DtcmPresent bool `xml:"dtcmPresent,omitempty"`

	// Indicate whether the Vector Table Offset Register (VTOR)
	// is implemented in Cortex-M0+ based devices.
	// If not specified, then VTOR is assumed to be present.
	VtorPresent bool `xml:"vtorPresent,omitempty"`

	// Define the number of bits available in the Nested Vectored
	// Interrupt Controller (NVIC) for configuring priority.
	NvicPrioBits string `xml:"nvicPrioBits"`

	// Indicate whether the processor implements a vendor-specific
	// System Tick Timer.
	// If false, then the Arm-defined System Tick Timer is available.
	// If true, then a vendor-specific System Tick Timer must be
	// implemented.
	VendorSystickConfig bool `xml:"vendorSystickConfig"`

	// Add 1 to the highest interrupt number and specify this number
	// in here.
	// You can start to enumerate interrupts from 0.
	// Gaps might exist between interrupts.
	// For example, you have defined interrupts with the numbers 1, 2, and 8.
	// Add 9 :(8+1) into this field.
	DeviceNumInterrupts uint `xml:"deviceNumInterrupts,omitempty"`

	// Indicate the amount of regions in the Security Attribution Unit (SAU).
	// If the value is greater than zero, then the device has a SAU and the
	// number indicates the maximum amount of available address regions.
	SauNumRegions uint `xml:"sauNumRegions,omitempty"`

	// If the Secure Attribution Unit is preconfigured by HW or
	// Firmware, then the settings are described here.
	SauRegionsConfig *SauRegionsConfigType `xml:"sauRegionsConfig,omitempty"`
}

func (cpu *Cpu) Select(name CpuName) {
	cpu.Name = name
	if name == CpuNameCM4 ||
		name == CpuNameCM7 ||
		name == CpuNameCM33 ||
		name == CpuNameCM35P {
		cpu.FpuPresent = true
	}
	if name == CpuNameCM7 {
		cpu.FpuDP = true
		cpu.IcachePresent = true
		cpu.DcachePresent = true
		cpu.ItcmPresent = true
		cpu.DtcmPresent = true
	}
}
