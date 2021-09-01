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
	Size string `xml:"size"`

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

type DataType string

const (
	// unsigned byte
	DataTypeUInt8 DataType = "uint8_t"
	// unsigned half word
	DataTypeUInt16 DataType = "uint16_t"
	// unsigned word
	DataTypeUInt32 DataType = "uint32_t"
	// unsigned double word
	DataTypeUInt64 DataType = "uint64_t"
	// signed byte
	DataTypeInt8 DataType = "int8_t"
	// signed half word
	DataTypeInt16 DataType = "int16_t"
	// signed word
	DataTypeInt32 DataType = "int32_t"
	// signed double word
	DataTypeInt64 DataType = "int64_t"
	// pointer to unsigned byte
	DataTypeUInt8P DataType = "uint8_t *"
	// pointer to unsigned half word
	DataTypeUInt16P DataType = "uint16_t *"
	// pointer to unsigned word
	DataTypeUInt32P DataType = "uint32_t *"
	// pointer to unsigned double word
	DataTypeUInt64P DataType = "uint64_t *"
	// pointer to signed byte
	DataTypeInt8P DataType = "int8_t *"
	// pointer to signed half word
	DataTypeInt16P DataType = "int16_t *"
	// pointer to signed word
	DataTypeInt32P DataType = "int32_t *"
	// pointer to signed double word
	DataTypeInt64P DataType = "int64_t *"
)

type ModifiedWriteValues string

const (
	// write data bits of one shall clear (set to zero) the
	// corresponding bit in the register.
	ModifiedWriteValuesOneToClear ModifiedWriteValues = "oneToClear"
	// write data bits of one shall set (set to one) the
	// corresponding bit in the register.
	ModifiedWriteValuesOneToSet ModifiedWriteValues = "oneToSet"
	// write data bits of one shall toggle (invert) the
	// corresponding bit in the register.
	ModifiedWriteValuesOneToToggle ModifiedWriteValues = "oneToToggle"
	// write data bits of zero shall clear (set to zero) the
	// corresponding bit in the register.
	ModifiedWriteValuesZeroToClear ModifiedWriteValues = "zeroToClear"
	// write data bits of zero shall set (set to one) the
	// corresponding bit in the register.
	ModifiedWriteValuesZeroToSet ModifiedWriteValues = "zeroToSet"
	// write data bits of zero shall toggle (invert) the
	// corresponding bit in the register.
	ModifiedWriteValuesZeroToToggle ModifiedWriteValues = "zeroToToggle"
	// after a write operation all bits in the field are
	// cleared (set to zero).
	ModifiedWriteValuesClear ModifiedWriteValues = "clear"
	// after a write operation all bits in the field are
	// set (set to one).
	ModifiedWriteValuesSet ModifiedWriteValues = "set"
	// after a write operation all bit in the field may be
	// modified (default).
	ModifiedWriteValuesModify ModifiedWriteValues = "modify"
)

type Range struct {
	// Specify the smallest number to be written to the field.
	Minimum uint `xml:"minimum"`

	// Specify the largest number to be written to the field.
	Maximum uint `xml:"maximum"`
}

// Define constraints for writing values to a field.
// You can choose between three options, which are mutualy exclusive.
type WriteConstraint struct {
	// If true, only the last read value can be written.
	WriteAsRead bool `xml:"writeAsRead,omitempty"`

	// If true, only the values listed in the enumeratedValues list
	// can be written.
	UseEnumeratedValues bool `xml:"useEnumeratedValues,omitempty"`

	Range *Range `xml:"range,omitempty"`
}

type ReadAction string

const (
	// The register is cleared (set to zero) following a read operation.
	ReadActionClear ReadAction = "clear"
	// The register is set (set to ones) following a read operation.
	ReadActionSet ReadAction = "set"
	// The register is modified in some way after a read operation.
	ReadActionModify ReadAction = "modify"
	// One or more dependent resources other than the current register
	// are immediately affected by a read operation (it is recommended
	// that the register description specifies these dependencies).
	ReadActionModifyExternal ReadAction = "modifyExternal"
)

// The concept of enumerated values creates a map between unsigned
// integers and an identifier string.
// In addition, a description string can be associated with each
// entry in the map.
//   0 <-> disabled -> "The clock source clk0 is turned off."
//   1 <-> enabled  -> "The clock source clk1 is running."
//   2 <-> reserved -> "Reserved values. Do not use."
//   3 <-> reserved -> "Reserved values. Do not use."
// This information generates an enum in the device header file.
// The debugger may use this information to display the identifier
// string as well as the description.
// Just like symbolic constants making source code more readable,
// the system view in the debugger becomes more instructive.
// The detailed description can provide reference manual level
// details within the debugger.
type EnumeratedValues struct {
	// Makes a copy from a previously defined enumeratedValues section.
	// No modifications are allowed.
	// An enumeratedValues entry is referenced by its name.
	// If the name is not unique throughout the description, it needs
	// to be further qualified by specifying the associated field,
	// register, and peripheral as required. For example:
	//   field:                           clk.dis_en_enum
	//   register + field:                ctrl.clk.dis_en_enum
	//   peripheral + register + field:   timer0.ctrl.clk.dis_en_enum
	DerivedFrom string `xml:"derivedFrom,attr,omitempty"`

	// Identifier for the whole enumeration section.
	Name string `xml:"name,omitempty"`

	// Identifier for the enumeration section.
	// Overwrites the hierarchical enumeration type in the device
	// header file.
	// User is responsible for uniqueness across description.
	HeaderEnumName string `xml:"headerEnumName,omitempty"`

	// Possible values are "read", "write", or "read-write".
	// This allows specifying two different enumerated values
	// depending whether it is to be used for a read or a write access.
	// If not specified, the default value read-write is used.
	Usage string `xml:"usage,omitempty"`

	// Describes a single entry in the enumeration. The number of
	// required items depends on the bit-width of the associated field.
	EnumeratedValue []EnumeratedValue `xml:"enumeratedValue"`
}

// A bit-field has a name that is unique within the register.
// The position and size within the register can be decsribed
// in two ways:
//   - by the combination of the least significant bit's position
//     (lsb) and the most significant bit's position (msb), or
//   - the lsb and the bit-width of the field.
// A field may define an enumeratedValue in order to make the
// display more intuitive to read.
type Field struct {
	// Specify the field name from which to inherit data.
	// Elements specified subsequently override inherited values.
	// Usage:
	// Always use the full qualifying path, which must start with the
	// peripheral <name>, when deriving from another scope.
	// (for example, in periperhal A and registerX, derive from
	// peripheralA.registerYY.fieldYY.
	// You can use the field <name> only when both fields are in the
	// same scope.
	// No relative paths will work.
	// Remarks: When deriving, it is mandatory to specify at least
	// the <name> and <description>.
	DerivedFrom string `xml:"derivedFrom,attr,omitempty"`

	// Defines the number of elements in a list.
	Dim string `xml:"dim,omitempty"`

	// 	Specify the address increment, in bits, between two neighboring
	// list members in the address map.
	DimIncrement string `xml:"dimIncrement,omitempty"`

	// Specify the strings that substitue the placeholder %s within
	// <name> and <displayName>.
	DimIndex DimIndex `xml:"dimIndex,omitempty"`

	// Specify the name of the C-type structure.
	// If not defined, then the entry in the <name> element is used.
	DimName DimName `xml:"dimName,omitempty"`

	// Grouping element to create enumerations in the header file.
	DimArrayIndex *DimArrayIndex `xml:"dimArrayIndex,omitempty"`

	// Name string used to identify the field.
	// Field names must be unique within a register.
	Name string `xml:"name"`

	// String describing the details of the register.
	Description string `xml:"description,omitempty"`

	// Three mutually exclusive options exist to describe the bit-range:
	// 1. bitRangeLsbMsbStyle
	//   Value defining the position of the least significant bit of
	//   the field within the register.
	BitOffset string `xml:"bitOffset,omitempty"`

	//   Value defining the bit-width of the bitfield within the register.
	BitWidth string `xml:"bitWidth,omitempty"`

	// 2. bitRangeOffsetWidthStyle
	//   Value defining the bit position of the least significant
	//   bit within the register.
	Lsb string `xml:"lsb,omitempty"`

	// 	 Value defining the bit position of the most significant
	//   bit within the register.
	Msb string `xml:"msb,omitempty"`

	// 3. bitRangePattern
	//   A string in the format: "[<msb>:<lsb>]"
	BitRange string `xml:"bitRange,omitempty"`

	// Predefined strings set the access type. The element can be omitted
	// if access rights get inherited from parent elements.
	Access *AccessType `xml:"access"`

	// Describe the manipulation of data written to a field. If not specified,
	// the value written to the field is the value stored in the field.
	ModifiedWriteValues *ModifiedWriteValues `xml:"modifiedWriteValues,omitempty"`

	// Three mutually exclusive options exist to set write-constraints.
	WriteConstraint *WriteConstraint `xml:"writeConstraint,omitempty"`

	// If set, it specifies the side effect following a read operation.
	// If not set, the field is not modified after a read.
	ReadAction *ReadAction `xml:"readAction,omitempty"`

	// Next lower level of description.
	EnumeratedValues *EnumeratedValues `xml:"enumeratedValues"`
}

// Grouping element to define bit-field properties of a register.
type Fields struct {
	// Define the bit-field properties of a register.
	Field []Field `xml:"field"`
}

// The description of registers is the most essential part of SVD.
// If the elements <size>, <access>, <resetValue>, and <resetMask>
// have not been specified on a higher level, then these elements are
// mandatory on register level.
// A register can represent a single value or can be subdivided into
// individual bit-fields of specific functionality and semantics.
// From a schema perspective, the element <fields> is optional, however,
// from a specification perspective, <fields> are mandatory when they
// are described in the device documentation.
// You can define register arrays where the single description gets
// duplicated automatically.
// The size of the array is specified by the <dim> element.
// Register names get composed by the element <name> and the
// index-specific string defined in <dimIndex>.
// The element <dimIncrement> specifies the address offset between
// two registers.
type Register struct {
	// Specify the register name from which to inherit data.
	// Elements specified subsequently override inherited values.
	// Usage:
	// Always use the full qualifying path, which must start with
	// the peripheral <name>, when deriving from another scope.
	// (for example, in periperhal B, derive from peripheralA.registerX).
	// You can use the register <name> when both registers are in the
	// same scope.
	// No relative paths will work.
	// Remarks: When deriving a register, it is mandatory to specify
	// at least the <name>, the <description>, and the <addressOffset>.
	DerivedFrom string `xml:"derivedFrom,attr,omitempty"`

	// Define the number of elements in an array of registers.
	// If <dimIncrement> is specified, this element becomes mandatory.
	Dim string `xml:"dim,omitempty"`

	// Specify the address increment, in Bytes, between two neighboring registers.
	DimIncrement string `xml:"dimIncrement,omitempty"`

	// Specify the substrings that replaces the %s placeholder within
	// name and displayName.
	// By default, the index is a decimal value starting with 0 for the
	// first register.
	// dimIndex should not be used together with the placeholder [%s],
	// but rather with %s.
	DimIndex DimIndex `xml:"dimIndex,omitempty"`

	// Specify the name of the C-type structure.
	// If not defined, then the entry of the <name> element is used.
	DimName DimName `xml:"dimName,omitempty"`

	// Grouping element to create enumerations in the header file.
	DimArrayIndex *DimArrayIndex `xml:"dimArrayIndex,omitempty"`

	// String to identify the register.
	// Register names are required to be unique within the scope
	// of a peripheral.
	// You can use the placeholder %s, which is replaced by the
	// dimIndex substring.
	// Use the placeholder [%s] only at the end of the identifier
	// to generate arrays in the header file.
	// The placeholder [%s] cannot be used together with dimIndex.
	Name string `xml:"name"`

	// When specified, then this string can be used by a graphical
	// frontend to visualize the register.
	// Otherwise the name element is displayed.
	// displayName may contain special characters and white spaces.
	// You can use the placeholder %s, which is replaced by the
	// dimIndex substring.
	// Use the placeholder [%s] only at the end of the identifier.
	// The placeholder [%s] cannot be used together with dimIndex.
	DisplayName string `xml:"displayName,omitempty"`

	// String describing the details of the register.
	Description string `xml:"description,omitempty"`

	// Specifies a group name associated with all alternate register
	// that have the same name.
	// At the same time, it indicates that there is a register definition
	// allocating the same absolute address in the address space.
	AlternateGroup string `xml:"alternateGroup,omitempty"`

	// This tag can reference a register that has been defined above
	// to current location in the description and that describes the
	// memory location already.
	// This tells the SVDConv's address checker that the redefinition
	// of this particular register is intentional.
	// The register name needs to be unique within the scope of the
	// current peripheral.
	// A register description is defined either for a unique address
	// location or could be a redefinition of an already described address.
	// In the latter case, the register can be either marked
	// alternateRegister and needs to have a unique name, or it can have
	// the same register name but is assigned to a register subgroup
	// through the tag alternateGroup (specified in version 1.0).
	AlternateRegister string `xml:"alternateRegister,omitempty"`

	// Define the address offset relative to the enclosing element.
	AddressOffset string `xml:"addressOffset"`

	// Defines the default bit-width of any register contained in
	// the device (implicit inheritance).
	Size string `xml:"size,omitempty"`

	// Defines the default access rights for all registers.
	Access AccessType `xml:"access,omitempty"`

	// Defines the protection rights for all registers.
	Protection string `xml:"protection,omitempty"`

	// Defines the default value for all registers at RESET.
	ResetValue string `xml:"resetValue,omitempty"`

	// Identifies which register bits have a defined reset value.
	ResetMask string `xml:"resetMask,omitempty"`

	// It can be useful to assign a specific native C datatype to a register.
	// This helps avoiding type casts. For example, if a 32 bit
	// register shall act as a pointer to a 32 bit unsigned data item,
	// then dataType can be set to "uint32_t *".
	DataType DataType `xml:"dataType,omitempty"`

	// Element to describe the manipulation of data written to a register.
	// If not specified, the value written to the field is the
	// value stored in the field.
	ModifiedWriteValues ModifiedWriteValues `xml:"modifiedWriteValues,omitempty"`

	// Three mutually exclusive options exist to set write-constraints.
	WriteConstraint *WriteConstraint `xml:"writeConstraint,omitempty"`

	// If set, it specifies the side effect following a read operation.
	// If not set, the register is not modified.
	// Debuggers are not expected to read this register location unless
	// explicitly instructed by the user.
	ReadAction ReadAction `xml:"readAction,omitempty"`

	// In case a register is subdivided into bit fields, it should
	// be reflected in the SVD description file to create bit-access
	// macros and bit-field structures in the header file.
	Fields *Fields `xml:"fields,omitempty"`
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
	DimArrayIndex *DimArrayIndex `xml:"dimArrayIndex,omitempty"`

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
	GroupName string `xml:"groupName,omitempty"`

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
	Registers *Registers `xml:"registers,omitempty"`
}
