package svd

// Specify the strings that substitue the placeholder %s
// within <name> and <displayName>.
// By default, <dimIndex> is a value starting with 0.
// Remark: Do not define <dimIndex> when using the
// placeholder [%s] in <name> or <displayName>.
type DimIndex string

// Specify the name of the C-type structure.
// If not defined, then the entry in the <name> element is used.
type DimName string

type EnumeratedValue struct {
	// 	String describing the semantics of the value.
	// Can be displayed instead of the value.
	Name string `xml:"name,omitempty"`

	// Extended string describing the value.
	Description string `xml:"description,omitempty"`

	// Defines the constant for the bit-field as decimal,
	// hexadecimal (0x...) or binary (0b... or #...) number.
	// E.g.:
	//    <value>15</value>
	//    <value>0xf</value>
	//    <value>0b1111</value>
	//    <value>#1111</value>
	// In addition the binary format supports 'do not care'
	// bits represented by x.
	// E.g. specifying value 14 and 15 as:
	//    <value>0b111x</value>
	//    <value>#111x</value>
	Value string `xml:"value,omitempty"`

	// Defines the name and description for all other values
	// that are not listed explicitly.
	IsDefault bool `xml:"isDefault,omitempty"`
}

type DimArrayIndex struct {
	// Specify the base name of enumerations.
	// Overwrites the hierarchical enumeration type in the device
	// header file.
	// User is responsible for uniqueness across description.
	// The headerfile generator uses the name of a peripheral or
	// cluster as the base name for enumeration types.
	// If <headerEnumName> element is specfied, then this string
	// is used.
	HeaderEnumName string `xml:"headerEnumName,omitempty"`

	// Specify the values contained in the enumeration.
	EnumeratedValue []EnumeratedValue `xml:"enumeratedValue"`
}
