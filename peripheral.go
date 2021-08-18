package svd

type UsageType string

const (
	UsageRegisters UsageType = "registers"
	UsageBuffer    UsageType = "buffer"
	UsageReserved  UsageType = "reserved"
)

type AddressBlock struct {
	// Specifies the start address of an address block
	// relative to the peripheral baseAddress.
	Offset uint `xml:"offset"`

	// Specifies the number of addressUnitBits being covered
	// by this address block.
	// The end address of an address block results from the sum
	// of baseAddress, offset, and (size - 1).
	Size uint `xml:"size"`

	// Define usage type
	Usage UsageType `xml:"usage"`

	// Set the protection level for an address block.
	Protection ProtectionType `xml:"protection,omitempty"`
}

type Interrupt struct {
	// The string represents the interrupt name.
	Name string `xml:"name"`

	// The string describes the interrupt.
	Description string `xml:"description,omitempty"`

	// Represents the enumeration index value associated to the interrupt.
	Value string `xml:"value"`
}

type EnumeratedValues struct {
	EnumeratedValue []EnumeratedValue `xml:"enumeratedValue"`
}

type Field struct {
	Name             string           `xml:"name"`
	Description      string           `xml:"description"`
	BitRange         string           `xml:"bitRange"`
	Access           string           `xml:"access"`
	EnumeratedValues EnumeratedValues `xml:"enumeratedValues"`
}

type Fields struct {
	Field []Field `xml:"field"`
}

type Register struct {
	// Specify the cluster name from which to inherit data.
	// Elements specified subsequently override inherited values.
	// Usage:
	// Always use the full qualifying path, which must start with
	// the peripheral <name>, when deriving from another scope.
	// (for example, in periperhal B, derive from peripheralA.clusterX).
	// You can use the cluster <name> when both clusters are in the
	// same scope.
	// No relative paths will work.
	// Remarks: When deriving a cluster, it is mandatory to specify
	// at least the <name>, the <description>, and the <addressOffset>.
	DerivedFrom   string `xml:"derivedFrom,attr,omitempty"`
	Name          string `xml:"name"`
	Description   string `xml:"description"`
	AddressOffset string `xml:"addressOffset"`
	Size          string `xml:"size"`
	Access        string `xml:"access"`
	ResetValue    string `xml:"resetValue"`
	ResetMask     string `xml:"resetMask"`
	Fields        Fields `xml:"fields"`
	Dim           string `xml:"dim"`
	DimIncrement  string `xml:"dimIncrement"`
}

type Registers struct {
	// Define the sequence of register clusters.
	// Cluster []Cluster `xml:"cluster,omitempty"`

	// Define the sequence of registers.
	Register []Register `xml:"register"`
}

type Peripheral struct {
	// Specify the peripheral name from which to inherit data.
	// Elements specified subsequently override inherited values.
	DerivedFrom string `xml:"derivedFrom,attr,omitempty"`

	// Define the number of elements in an array.
	Dim uint `xml:"dim,omitempty"`

	// Specify the address increment, in Bytes, between two
	// neighboring array members in the address map.
	DimIncrement uint `xml:"dimIncrement,omitempty"`

	// Do not define on peripheral level.
	// By default, <dimIndex> is an integer value starting at 0.
	DimIndex DimIndex `xml:"dimIndex,omitempty"`

	// Specify the name of the C-type structure.
	// If not defined, then the entry of the <name> element is used.
	DimName DimName `xml:"dimName,omitempty"`

	// Grouping element to create enumerations in the header file.
	DimArrayIndex DimArrayIndex `xml:"dimArrayIndex,omitempty"`

	// The string identifies the peripheral.
	// Peripheral names are required to be unique for a device.
	// The name needs to be an ANSI C identifier to generate the header file.
	// You can use the placeholder [%s] to create arrays.
	Name string `xml:"name"`

	// The string specifies the version of this peripheral description.
	Version string `xml:"version,omitempty"`

	// The string provides an overview of the purpose and functionality
	// of the peripheral.
	Description string `xml:"description,omitempty"`

	// All address blocks in the memory space of a device are assigned
	// to a unique peripheral by default.
	// If multiple peripherals describe the same address blocks,
	// then this needs to be specified explicitly.
	// A peripheral redefining an address block needs to specify the
	// name of the peripheral that is listed first in the description.
	AlternatePeripheral string `xml:"alternatePeripheral,omitempty"`

	// Define a name under which the System Viewer is showing this peripheral.
	GroupName string `xml:"groupName"`

	// 	Define a string as prefix.
	// All register names of this peripheral get this prefix.
	PrependToName string `xml:"prependToName,omitempty"`

	// Define a string as suffix.
	// All register names of this peripheral get this suffix.
	AppendToName string `xml:"appendToName,omitempty"`

	// Specify the base name of C structures.
	// The headerfile generator uses the name of a peripheral as the
	// base name for the C structure type.
	// If <headerStructName> element is specfied, then this string
	// is used instead of the peripheral name;
	// useful when multiple peripherals get derived and a generic
	// type name should be used.
	HeaderStructName string `xml:"headerStructName,omitempty"`

	// Define a C-language compliant logical expression returning a
	// TRUE or FALSE result.
	// If TRUE, refreshing the display for this peripheral is disabled
	// and related accesses by the debugger are suppressed.
	// Only constants and references to other registers contained in
	// the description are allowed: <peripheral>-><register>-><field>,
	// for example, (System->ClockControl->apbEnable == 0).
	// The following operators are allowed in the expression
	// [&&,||, ==, !=, >>, <<, &, |].
	// Attention
	// Use this feature only in cases where accesses from the debugger
	// to registers of un-clocked peripherals result in severe
	// debugging failures.
	// SVD is intended to provide static information and does not
	// include any run-time computation or functions.
	// Such capabilities can be added by the tools,
	// and is beyond the scope of this description language.
	DisableCondition string `xml:"disableCondition,omitempty"`

	// Lowest address reserved or used by the peripheral.
	BaseAddress string `xml:"baseAddress"`

	// Define the default bit-width of any register contained in
	// the device (implicit inheritance).
	Size uint `xml:"size,omitempty"`

	// Define default access rights for all registers.
	Access AccessType `xml:"access,omitempty"`

	// Default protection rights for all registers.
	Protection ProtectionType `xml:"protection,omitempty"`

	// Default value for all registers at RESET.
	ResetValue uint `xml:"resetValue,omitempty"`

	// Define which register bits have a defined reset value.
	ResetMask uint `xml:"resetMask,omitempty"`

	// Specify an address range uniquely mapped to this peripheral.
	// A peripheral must have at least one address block,
	// but can allocate multiple distinct address ranges.
	// If a peripheral is derived from another peripheral,
	// the addressBlock is not mandatory.
	AddressBlock []AddressBlock `xml:"addressBlock,omitempty"`

	// A peripheral can have multiple associated interrupts.
	// This entry allows the debugger to show interrupt names
	// instead of interrupt numbers.
	Interrupt []Interrupt `xml:"interrupt,omitempty"`

	// Group to enclose register definitions.
	Registers Registers `xml:"registers,omitempty"`
}
