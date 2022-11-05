package svd

import (
	"reflect"
	"testing"
)

func TestNewDevice(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Device
	}{
		{
			name: "New 1",
			args: args{
				name: "DeviceName",
			},
			want: &Device{
				SchemaVersion:             "1.3.6",
				Xs:                        "http://www.w3.org/2001/XMLSchema-instance",
				NoNamespaceSchemaLocation: "CMSIS-SVD.xsd",
				Name:                      "DeviceName",
				AddressUnitBits:           8,
				Width:                     32,
				Size:                      32,
				Access:                    AccessReadWrite,
				ResetValue:                "0x00000000",
				ResetMask:                 "0xFFFFFFFF",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDevice(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_SVD(t *testing.T) {
	tests := []struct {
		name    string
		dev     Device
		wantSvd []byte
		wantErr bool
	}{
		{
			name:    "SVD minimal",
			dev:     Device{},
			wantSvd: []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<device xmlns:xs=\"\" xs:noNamespaceSchemaLocation=\"\" schemaVersion=\"\">\n  <name></name>\n  <version></version>\n  <description></description>\n  <cpu>\n    <name></name>\n    <revision></revision>\n    <endian></endian>\n    <mpuPresent>false</mpuPresent>\n    <fpuPresent>false</fpuPresent>\n    <nvicPrioBits></nvicPrioBits>\n    <vendorSystickConfig>false</vendorSystickConfig>\n  </cpu>\n  <addressUnitBits>0</addressUnitBits>\n  <width>0</width>\n  <peripherals></peripherals>\n</device>"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSvd, err := tt.dev.SVD()
			if (err != nil) != tt.wantErr {
				t.Errorf("Device.SVD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSvd, tt.wantSvd) {
				t.Errorf("Device.SVD() = %v, want %v", gotSvd, tt.wantSvd)
			}
		})
	}
}
