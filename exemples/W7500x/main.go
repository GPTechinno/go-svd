package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/GPTechinno/go-svd"
	// "github.com/davecgh/go-spew/spew"
)

func str2dec(hex string) uint64 {
	i, _ := strconv.ParseUint(hex, 10, 64)
	return i
}

func hex2dec(hex string) uint64 {
	i, _ := strconv.ParseUint(hex, 16, 64)
	return i
}

func dec2hex(dec uint64) string {
	return strconv.FormatUint(dec, 16)
}

func dec2str(dec uint64) string {
	return strconv.FormatUint(dec, 10)
}

func hex2dec2hex(hex string) string {
	return dec2hex(hex2dec(hex))
}

func main() {
	pInput := flag.String("i", "", "CMSIS Device header file (w7500x.h)")
	flag.Parse()
	if *pInput == "" {
		log.Fatalln("w7500x.h is mandatory (from W7500x_StdPeriph_Lib/Libraries/CMSIS/Device/WIZnet/W7500/Include/w7500x.h)")
	}
	// Source file 1 : CMSIS Device header
	dat, err := ioutil.ReadFile(*pInput)
	if err != nil {
		log.Fatalln(err)
	}
	w7500xH := string(dat)

	// Device
	dev := svd.NewDevice("W7500x")
	dev.Vendor = "WIZnet"
	dev.Description = "The IOP (Internet Offload Processor) W7500P is the one-chip solution which integrates an ARM Cortex-M0, 128KB Flash and hardwired TCP/IP core & PHY for various embedded application platform especially requiring ‘Internet of things’.\nThe TCP/IP core is a market-proven hardwired TCP/IP stack with an integrated Ethernet MAC. The Hardwired TCP/IP stack supports the TCP, UDP, IPv4, ICMP, ARP, IGMP and PPPoE which has been used in various applications for years. W7500P suits best for users who need Internet connectivity for application."
	// Version
	verMain := regexp.MustCompile(`#define[\s]__W7500X_STDPERIPH_VERSION_MAIN[\s]+\(0x([0-9a-fA-F]{2})\)`).FindStringSubmatch(w7500xH)
	verSub1 := regexp.MustCompile(`#define[\s]__W7500X_STDPERIPH_VERSION_SUB1[\s]+\(0x([0-9a-fA-F]{2})\)`).FindStringSubmatch(w7500xH)
	verSub2 := regexp.MustCompile(`#define[\s]__W7500X_STDPERIPH_VERSION_SUB2[\s]+\(0x([0-9a-fA-F]{2})\)`).FindStringSubmatch(w7500xH)
	dev.Version = hex2dec2hex(verMain[1]) + "." + hex2dec2hex(verSub1[1]) + "." + hex2dec2hex(verSub2[1])

	// CPU
	dev.Cpu.Select(svd.CpuNameCM0)
	dev.Cpu.Endian = svd.EndianLittle
	// Core Revision
	coreRevs := regexp.MustCompile(`#define[\s]__CM0_REV[\s]+0x([0-9a-fA-F]{2})([0-9a-fA-F]{2})`).FindStringSubmatch(w7500xH)
	dev.Cpu.Revision = "r" + hex2dec2hex(coreRevs[1]) + "p" + hex2dec2hex(coreRevs[2])

	// NVIC Prio Bits
	dev.Cpu.NvicPrioBits = regexp.MustCompile(`#define[\s]__NVIC_PRIO_BITS[\s]+([0-9])`).FindStringSubmatch(w7500xH)[1]

	// Interrupts
	its := regexp.MustCompile(`[\s]+([0-9A-Za-z_/]+)_IRQn[\s]=[\s]([0-9]+),[\s]+/\*\!<\s([0-9A-Za-z_/\s]+)[\s]Interrupt[\s]+\*/`).FindAllStringSubmatch(w7500xH, -1)
	// store them for futur match within peripherals
	interrupts := make(map[string]svd.Interrupt)
	for _, it := range its {
		key := it[1]
		if key == "PORT0" {
			key = "GPIOA"
		}
		if key == "PORT1" {
			key = "GPIOB"
		}
		if key == "PORT2" {
			key = "GPIOC"
		}
		if key == "PORT3" {
			key = "GPIOD"
		}
		interrupts[key] = svd.Interrupt{Name: it[1], Description: it[3] + " Interrupt", Value: it[2]}
	}

	// Peripherals
	// store base addresses
	bases := make(map[string]uint64)
	// first raw base addresses
	rawBases := regexp.MustCompile(`#define[\s]([0-9A-Z]+_BASE)[\s]+\(0x([0-9a-fA-F]{8})UL\)`).FindAllStringSubmatch(w7500xH, -1)
	for _, b := range rawBases {
		bases[b[1]] = hex2dec(b[2])
	}
	// then linked base addresses
	linkedBases := regexp.MustCompile(`#define[\s]([0-9A-Z]+_BASE)[\s]+([0-9A-Z]+_BASE)`).FindAllStringSubmatch(w7500xH, -1)
	for _, b := range linkedBases {
		bases[b[1]] = bases[b[2]]
	}
	// then offset base addresses
	offsetBases := regexp.MustCompile(`#define[\s]([0-9A-Z]+_(BASE|OSC|BGT))[\s]+\(([0-9A-Z]+_BASE)[\s]\+[\s]0x([0-9a-fA-F]{8})UL\)`).FindAllStringSubmatch(w7500xH, -1)
	for _, b := range offsetBases {
		bases[b[1]] = bases[b[3]] + hex2dec(b[4])
	}

	// Peripherals
	periphs := regexp.MustCompile(`#define[\s]([0-9A-Z\_]+)[\s]+\(\(([0-9A-Za-z\_]+_TypeDef)[\s]\*\)[\s]+([0-9A-Zx\_\s\+\(\)]+)\)`).FindAllStringSubmatch(w7500xH, -1)
	currentTypeDef := ""
	referenceName := ""
	for _, p := range periphs {
		// log.Println(p[1])
		// GroupName
		grpName := regexp.MustCompile(`[0-9]`).ReplaceAllString(strings.Split(p[1], "_")[0], "")
		if len(grpName) == 5 && grpName[:4] == "GPIO" {
			grpName = "GPIO"
		}
		// BaseAddress
		var baseAdd uint64
		if p[3][0] == '(' {
			add := regexp.MustCompile(`\(([0-9A-Z]+_BASE)\s\+\s0x([0-9a-fA-F]+)UL\)`).FindStringSubmatch(p[3])
			baseAdd = bases[add[1]] + hex2dec(add[2])
		} else {
			baseAdd = bases[p[3]]
		}
		// Construct Peripheral
		peripheral := svd.Peripheral{
			Name:        p[1],
			BaseAddress: "0x" + dec2hex(baseAdd),
		}
		// Interrupt
		key := p[1]
		if len(key) == 12 && key[:9] == "DUALTIMER" {
			key = key[:10]
		}
		if it, ok := interrupts[key]; ok {
			peripheral.Interrupt = append(peripheral.Interrupt, it)
			if key == p[1] || p[1][11] == '1' {
				delete(interrupts, key)
			}
		}
		if currentTypeDef == p[2] {
			peripheral.DerivedFrom = referenceName
		} else {
			currentTypeDef = p[2]
			referenceName = p[1]
			peripheral.GroupName = grpName
			// Specific TypeDef structure
			regs, addBlock := getRegisters(w7500xH, p[2])
			reg := svd.Registers{
				Register: regs,
			}
			peripheral.Registers = &reg
			peripheral.AddressBlock = append(peripheral.AddressBlock, addBlock)
		}
		// Save Peripheral
		dev.Peripherals.Peripheral = append(dev.Peripherals.Peripheral, peripheral)
	}

	// if len(interrupts) > 0 {
	// 	log.Println("Left over Interrupt :")
	// 	spew.Dump(interrupts)
	// }

	// Generate SVD
	svd, err := dev.SVD()
	if err != nil {
		log.Fatalln(err)
	}

	os.Stdout.Write(svd)
}

func getRegisters(file, typeDefName string) (regs []svd.Register, addBlocks svd.AddressBlock) {
	// first fix error in the source file !!!
	file = strings.ReplaceAll(file, "#define ADC_CHSEL_CHSEL                 (0x0UL)", "#define ADC_CHSEL_CHSEL                 (0xFUL)")
	file = strings.ReplaceAll(file, "ADC_CTR_PWD_PWD", "ADC_CTR_PWD")
	file = strings.ReplaceAll(file, "ADC_CTR_PWD_SMPSEL", "ADC_CTR_SMPSEL")
	file = strings.ReplaceAll(file, "CRG_RTC_SSR_RTCHS", "CRG_RTC_SSR_RTCSEL")
	file = strings.ReplaceAll(file, "#define CRG_MONCLK_SSR_CLKMON_SEL       (0x00UL)", "#define CRG_MONCLK_SSR_CLKMON_SEL       (0x1FUL)")
	// missing WDOGCLK_SSR in CRG
	// file = strings.ReplaceAll(file, "	    __IO uint32_t WDOGCLK_HS_PVSR;      /*!< WDOGCLK High Speed prescale value select register,                 Address offset : 0x144 */", "	    __IO uint32_t WDOGCLK_HS_PVSR;      /*!< WDOGCLK High Speed prescale value select register,                 Address offset : 0x144 */\r\n	    __IO uint32_t WDOGCLK_SSR;      /*!< WDOGCLK clock source select register,                 Address offset : 0x14C */")
	file = strings.ReplaceAll(file, "CRG_UARTCLK_PVSR_UCP            (0x00UL)", "CRG_UARTCLK_PVSR_UCP            (0x03UL)")
	file = strings.ReplaceAll(file, "DMA_STATUS_CR", "DMA_STATUS_STATE")
	file = strings.ReplaceAll(file, "DMA_WAITONREQ_STATUS", "DMA_WAITONREQ_STATUS_DMA_WAITONREQ")
	file = strings.ReplaceAll(file, "#define DMA_ERR_CLR                     (0x3FUL)            /*!< ERR_CLR[5:0] bits (Returns the status of DMA_ERR, or set the signal LOW) */", "#define DMA_ERR_CLR                     (0x01UL)            /*!< Returns the status of DMA_ERR, or set the signal LOW */")
	file = strings.ReplaceAll(file, "S_UART_DR_DATA                  (0xFFUL)", "S_UART_DR_DATA (0xFFUL) /*!< Receive (READ)/Transmit (WRITE) data */")
	file = strings.ReplaceAll(file, "UARTR_SR", "UART_RSR")
	file = strings.ReplaceAll(file, "UARTR_CR", "UART_ECR")
	file = strings.ReplaceAll(file, "/*!< Alternate Function Set register,       Address offset : 0x018 */", "")
	file = strings.ReplaceAll(file, "/*!< Alternate Function Clear register,     Address offset : 0x01C */", "")
	td := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSuffix(typeDefName, "_TypeDef"), "PWM", "PWM_CHn"), "PWM_CHn_Common", "PWM_CM")
	typeDef := findTypeDef(file, typeDefName)
	registers := regexp.MustCompile(`__(I|O|IO)[\s]+uint32_t\s([0-9A-Z\_]+)[\[]?([0-9]*)[\]]?;[\s]+/\*!<\s([0-9a-zA-Z\s\(\)\:\_\-/]+),[\s]+Address\soffset[\s]?:\s(0x[0-9A-F]+)`).FindAllStringSubmatch(typeDef, -1)
	var highestOffset uint64
	var lastElementSize uint64
	for _, r := range registers {
		if hex2dec(r[5][2:]) > highestOffset {
			highestOffset = hex2dec(r[5][2:])
		}
		lastElementSize = 4
		reg := svd.Register{
			Name:          r[2],
			Description:   strings.TrimSpace(r[4]),
			AddressOffset: r[5],
			ResetValue:    getResetValue(td, r[2]),
		}
		if r[3] != "" {
			reg.Name += "[%s]"
			reg.Dim = r[3]
			reg.DimIncrement = "4"
			lastElementSize *= str2dec(r[3])
		}
		if r[1] == "I" {
			reg.Access = svd.AccessReadOnly
		} else if r[1] == "O" {
			reg.Access = svd.AccessWriteOnly
			// not needed it is default
			// } else if r[1] == "IO" {
			// 	reg.Access = svd.AccessReadWrite
		}
		// Fields
		fies := svd.Fields{}
		regName := r[2]
		if td == "CRG" {
			if len(regName) > 6 && regName[:5] == "TIMER" {
				regName = "TIMERCLK" + regName[9:]
			}
			if len(regName) > 4 && regName[:3] == "PWM" {
				regName = "PWMCLK" + regName[7:]
			}
		}
		special := ""
		if td == "UART" && regName == "CR" {
			special = "nut12"
		}
		re := `#define\s` + td + `_` + regName + `_([A-Z` + special + `\_]+)[\s]+\(0x([0-9A-F]+)[U]?[L]?\)[\s]+/\*!<\s([0-9A-Za-z\s\[\:\]\(\)\-\_\,\/\.]+)\*/`
		fields := regexp.MustCompile(re).FindAllStringSubmatch(file, -1)
		if len(fields) == 0 {
			re = `#define\s` + td + `_` + regName + `()[\s]+\(0x([0-9A-F]+)[U]?[L]?\)[\s]+/\*!<\s([0-9A-Za-z\s\[\:\]\(\)\-\_\,\/\.]+)\*/`
			fields = regexp.MustCompile(re).FindAllStringSubmatch(file, -1)
		}
		if len(fields) == 0 {
			re = `#define\s` + td + `_` + regName + `_([A-Z0-9\_]+)[\s]+\(0x([0-9A-F]+)[U]?[L]?\)[\s]+/\*!<\s([0-9A-Za-z\s\[\:\]\(\)\-\_\,\/\.]+)\*/`
			fields = regexp.MustCompile(re).FindAllStringSubmatch(file, -1)
		}
		var resetMask uint64
		for _, f := range fields {
			d := regexp.MustCompile(`\[([0-9]+):0\]\sbits\s\(([A-Za-z0-9\s\-]+)\)`).FindStringSubmatch(f[3])
			fi := svd.Field{
				Name:     f[1],
				BitRange: "[" + getMsb(hex2dec(f[2])) + ":" + getLsb(hex2dec(f[2])) + "]",
			}
			resetMask += hex2dec(f[2])
			if fi.Name == "" {
				fi.Name = reg.Name
			}
			if len(d) > 0 {
				fi.Description = strings.ReplaceAll(d[2], "  ", " ")
			} else {
				fi.Description = strings.ReplaceAll(strings.TrimSpace(f[3]), "  ", " ")
			}
			fi.EnumeratedValues = getEnumeratedValues(td, r[2], fi.Name)
			fies.Field = append(fies.Field, fi)
		}
		if len(fies.Field) > 0 {
			reg.Fields = &fies
			// } else {
			// 	log.Println(re, len(fies.Field))
		}
		reg.ResetMask = "0x" + dec2hex(resetMask)
		// Add it to the list
		regs = append(regs, reg)
	}
	addBlocks.Size = "0x" + dec2hex(highestOffset+lastElementSize)
	addBlocks.Usage = svd.UsageRegisters
	return
}

func findTypeDef(file, typeDefName string) (typeDef string) {
	scanner := bufio.NewScanner(strings.NewReader(file))
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if found {
			typeDef += line + "\n"
		}
		if line == "typedef struct" {
			found = true
		}
		if len(line) > 1 && line[0] == '}' {
			found = false
			if line == "} "+typeDefName+";" {
				return
			}
			typeDef = ""
		}
	}
	return
}

func getLsb(mask uint64) string {
	for i := 0; i < 32; i++ {
		if mask&(1<<i) > 0 {
			return dec2str(uint64(i))
		}
	}
	return "not found"
}

func getMsb(mask uint64) string {
	for i := 31; i >= 0; i-- {
		if mask&(1<<i) > 0 {
			return dec2str(uint64(i))
		}
	}
	return "not found"
}

// Manual Input according to W7500x Reference Manual Version 1.1.0
func getResetValue(typeDefName, regName string) string {
	notDefault := make(map[string]string)
	notDefault["CRG:PLL_PDR"] = "0x00000001"
	notDefault["CRG:PLL_FCR"] = "0x00050200"
	notDefault["CRG:PLL_OER"] = "0x00000001"
	notDefault["CRG:FCLK_SSR"] = "0x00000001"
	notDefault["CRG:SSPCLK_SSR"] = "0x00000001"
	notDefault["CRG:ADCCLK_SSR"] = "0x00000001"
	notDefault["CRG:TIMER0CLK_SSR"] = "0x00000001"
	notDefault["CRG:TIMER1CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM0CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM1CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM2CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM3CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM4CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM5CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM6CLK_SSR"] = "0x00000001"
	notDefault["CRG:PWM7CLK_SSR"] = "0x00000001"
	notDefault["CRG:RTC_HS_SSR"] = "0x00000001"
	notDefault["CRG:WDOGCLK_HS_SSR"] = "0x00000001"
	notDefault["CRG:UARTCLK_SSR"] = "0x00000001"
	notDefault["CRG:MIICLK_ECR"] = "0x00000003"
	notDefault["RNG:POLY"] = "0xE0000202"
	notDefault["DMA:STATUS"] = "0x00050000" // ???
	notDefault["ADC:CTR"] = "0x00000003"
	notDefault["WDG:LOAD"] = "0xFFFFFFFF"
	notDefault["WDG:VALUE"] = "0xFFFFFFFF"
	notDefault["UART:FR"] = "0bx11000xxx"
	notDefault["UART:CR"] = "0x0300"
	notDefault["UART:IFLS"] = "0x12"
	if v, ok := notDefault[typeDefName+":"+regName]; ok {
		return v
	}
	return ""
}

// Manual Input according to W7500x Reference Manual Version 1.1.0
func getEnumeratedValues(typeDefName, regName, fieldName string) *svd.EnumeratedValues {
	notDefault := make(map[string]svd.EnumeratedValues)
	notDefault["CRG:OSC_PDR:OSCPD"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Normal"},
		{Value: "1", Name: "PowerDown"},
	}}
	notDefault["CRG:PLL_PDR:PLLPD"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "PowerDown"},
		{Value: "1", Name: "Normal"},
	}}
	notDefault["CRG:PLL_OER:PLLOEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable", Description: "VCO is working but FOUT is low only"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["CRG:PLL_BPR:PLLBPN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable", Description: "Normal operation"},
		{Value: "1", Name: "Enable", Description: "Clock out is clock input"},
	}}
	notDefault["CRG:PLL_IFSR:PLLIS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "1", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:FCLK_SSR:FCKSRC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{IsDefault: true, Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:FCLK_PVSR:FCKPRE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Bypass", Description: "1/1"},
		{Value: "0b01", Name: "Half", Description: "1/2"},
		{Value: "0b10", Name: "By4", Description: "1/4"},
		{Value: "0b11", Name: "By8", Description: "1/8"},
	}}
	notDefault["CRG:SSPCLK_SSR:SSPCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:SSPCLK_PVSR:SSPCP"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Bypass", Description: "1/1"},
		{Value: "0b01", Name: "Half", Description: "1/2"},
		{Value: "0b10", Name: "By4", Description: "1/4"},
		{Value: "0b11", Name: "By8", Description: "1/8"},
	}}
	notDefault["CRG:ADCCLK_SSR:ADCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:ADCCLK_PVSR:ADCCP"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Bypass", Description: "1/1"},
		{Value: "0b01", Name: "Half", Description: "1/2"},
		{Value: "0b10", Name: "By4", Description: "1/4"},
		{Value: "0b11", Name: "By8", Description: "1/8"},
	}}
	notDefault["CRG:TIMER0CLK_SSR:TCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:TIMER0CLK_PVSR:TCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:TIMER1CLK_SSR:TCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:TIMER1CLK_PVSR:TCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM0CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM0CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM1CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM1CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM2CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM2CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM3CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM3CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM4CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM4CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM5CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM5CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM6CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM6CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:PWM7CLK_SSR:PCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:PWM7CLK_PVSR:PCPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:RTC_HS_SSR:RTCHS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:RTC_HS_PVSR:RTCPRE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:RTC_SSR:RTCSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "RTCCLK_hs"},
		{Value: "1", Name: "32K_OSC_CLK", Description: "Low speed external oscillator clock"},
	}}
	notDefault["CRG:WDOGCLK_HS_SSR:WDHS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:WDOGCLK_HS_PVSR:WDPRE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Bypass", Description: "1/1"},
		{Value: "0b001", Name: "Half", Description: "1/2"},
		{Value: "0b010", Name: "By4", Description: "1/4"},
		{Value: "0b011", Name: "By8", Description: "1/8"},
		{Value: "0b100", Name: "By16", Description: "1/16"},
		{Value: "0b101", Name: "By32", Description: "1/32"},
		{Value: "0b110", Name: "By64", Description: "1/64"},
		{Value: "0b111", Name: "By128", Description: "1/128"},
	}}
	notDefault["CRG:WDOGCLK_SSR:WDSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "WDOGCLK_hs"},
		{Value: "1", Name: "32K_OSC_CLK", Description: "Low speed external oscillator clock"},
	}}
	notDefault["CRG:UARTCLK_SSR:UCSS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Disable"},
		{Value: "0b01", Name: "PLL", Description: "Output clock of PLL (MCLK)"},
		{Value: "0b10", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b11", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
	}}
	notDefault["CRG:UARTCLK_PVSR:UCP"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Bypass", Description: "1/1"},
		{Value: "0b01", Name: "Half", Description: "1/2"},
		{Value: "0b10", Name: "By4", Description: "1/4"},
		{Value: "0b11", Name: "By8", Description: "1/8"},
	}}
	notDefault["CRG:MIICLK_ECR:MIIREN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["CRG:MIICLK_ECR:MIITEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["CRG:MONCLK_SSR:CLKMON_SEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00000", Name: "PLL", Description: "MCLK"},
		{Value: "0b00001", Name: "FCLK"},
		{Value: "0b00010", Name: "Internal", Description: "Internal 8MHz RC oscillator clock (RCLK)"},
		{Value: "0b00011", Name: "External", Description: "External oscillator clock (OCLK, 8MHz-24MHz)"},
		{Value: "0b00100", Name: "ADCCLK"},
		{Value: "0b00101", Name: "SSPCLK"},
		{Value: "0b00110", Name: "TIMCLK0"},
		{Value: "0b00111", Name: "TIMCLK1"},
		{Value: "0b01000", Name: "PWMCLK0"},
		{Value: "0b01001", Name: "PWMCLK1"},
		{Value: "0b01010", Name: "PWMCLK2"},
		{Value: "0b01011", Name: "PWMCLK3"},
		{Value: "0b01100", Name: "PWMCLK4"},
		{Value: "0b01101", Name: "PWMCLK5"},
		{Value: "0b01110", Name: "PWMCLK6"},
		{Value: "0b01111", Name: "PWMCLK7"},
		{Value: "0b10000", Name: "UARTCLK"},
		{Value: "0b10001", Name: "MII_RCK"},
		{Value: "0b10010", Name: "MII_TCK"},
		{Value: "0b10011", Name: "RTCCLK"},
	}}
	notDefault["RNG:RUN:RUN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Run"},
	}}
	notDefault["RNG:CLKSEL:CLKSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "RNG", Description: "refer to clock generator block"},
		{Value: "1", Name: "PCLK"},
	}}
	notDefault["RNG:MODE:MODE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "PLL_LOCK", Description: "which is for power on random number"},
		{Value: "1", Name: "RNG_RUN"},
	}}
	notDefault["DMA:STATUS:ENABLE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["DMA:STATUS:STATE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b0000", Name: "Idle"},
		{Value: "0b0001", Name: "ReadingChanCtrlData"},
		{Value: "0b0010", Name: "ReadingSrcDataEndPtr"},
		{Value: "0b0011", Name: "ReadingDstDataEndPtr"},
		{Value: "0b0100", Name: "ReadingSrcData"},
		{Value: "0b0101", Name: "WritingDstData"},
		{Value: "0b0110", Name: "WritingChanCtrlData"},
		{Value: "0b1000", Name: "Stalled"},
		{Value: "0b1001", Name: "Done"},
		{Value: "0b1010", Name: "PeriphScatGathTrans"},
	}}
	notDefault["DMA:CFG:ENABLE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["DMA:ERR_CLR:ERR_CLR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Low"},
		{Value: "1", Name: "High"},
	}}
	notDefault["ADC:CTR:SMPSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Abnormal"},
		{Value: "1", Name: "Normal"},
	}}
	notDefault["ADC:CTR:PWD"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Active"},
		{Value: "1", Name: "PowerDown"},
	}}
	notDefault["ADC:CHSEL:CHSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b0000", Name: "Channel0"},
		{Value: "0b0001", Name: "Channel1"},
		{Value: "0b0010", Name: "Channel2"},
		{Value: "0b0011", Name: "Channel3"},
		{Value: "0b0100", Name: "Channel4"},
		{Value: "0b0101", Name: "Channel5"},
		{Value: "0b0110", Name: "Channel6"},
		{Value: "0b0111", Name: "Channel7"},
		{Value: "0b1000", Name: "NoChannel"},
		{Value: "0b1111", Name: "LDOOutput1V5"},
	}}
	notDefault["ADC:START:ADC_SRT"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Ready"},
		{Value: "1", Name: "Start", Description: "This bit clear automatically after conversion"},
	}}
	notDefault["ADC:INT:INT"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Done"},
		{Value: "1", Name: "NotDone"},
	}}
	notDefault["ADC:INT:MASK"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["ADC:INTCLR:INTCLR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["PWM_CHn:IR:MI"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotOccured"},
		{Value: "1", Name: "Occured"},
	}}
	notDefault["PWM_CHn:IR:OI"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotOccured"},
		{Value: "1", Name: "Occured"},
	}}
	notDefault["PWM_CHn:IR:CI"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotOccured"},
		{Value: "1", Name: "Occured"},
	}}
	notDefault["PWM_CHn:IER:MIE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotEnabled"},
		{Value: "1", Name: "Enabled"},
	}}
	notDefault["PWM_CHn:IER:OIE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotEnabled"},
		{Value: "1", Name: "Enabled"},
	}}
	notDefault["PWM_CHn:IER:CIE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotEnabled"},
		{Value: "1", Name: "Enabled"},
	}}
	notDefault["PWM_CHn:ICR:MIC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["PWM_CHn:ICR:OIC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["PWM_CHn:ICR:CIC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["PWM_CHn:UDMR:UDM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Up"},
		{Value: "1", Name: "Down"},
	}}
	notDefault["PWM_CHn:TCMR:TCM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "Timer"},
		{Value: "0b01", Name: "CounterRising"},
		{Value: "0b10", Name: "CounterFalling"},
		{Value: "0b11", Name: "CounterToggle"},
	}}
	notDefault["PWM_CHn:PEEER:PEEE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "OutDisableInDisable"},
		{Value: "0b01", Name: "OutDisableInEnable"},
		{Value: "0b10", Name: "OutEnableInDisable"},
		// {Value: "0b11", Name: "OutEnableInEnable"}, // ???
	}}
	notDefault["PWM_CHn:CMR:CM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "RisingEdge"},
		{Value: "1", Name: "FallingEdge"},
	}}
	notDefault["PWM_CHn:PDMR:PDM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Periodic"},
		{Value: "1", Name: "OneShot"},
	}}
	notDefault["PWM_CHn:DZER:DZE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE0"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE1"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE2"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE3"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE4"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE5"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE6"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:IER:IE7"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["PWM_CM:SSR:SS0"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS1"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS2"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS3"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS4"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS5"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS6"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:SSR:SS7"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Stop"},
		{Value: "1", Name: "Start"},
	}}
	notDefault["PWM_CM:PSR:PS0"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS1"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS2"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS3"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS4"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS5"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS6"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["PWM_CM:PSR:PS7"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotPaused"},
		{Value: "1", Name: "Paused"},
	}}
	notDefault["WDT:CONTROL:IEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["WDT:CONTROL:REN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["WDT:INTCLR:WIC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["WDT:LOCK:WES"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "NotLocked"},
		{Value: "1", Name: "Locked"},
	}}
	notDefault["WDT:LOCK:ERW"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Lock"},
		{Value: "0x1ACCE551", Name: "UnLock"},
	}}
	notDefault["RTC:RTCCON:CLKEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCCON:DIVRST"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Reset"},
	}}
	notDefault["RTC:RTCCON:INTEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:IMSEC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:IMMIN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:IMHOUR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:IMDATE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:IMDAY"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:IMMON"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTE:AINT"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCINTP:RTCCIF"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["RTC:RTCINTP:RTCALF"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["RTC:RTCAMR:AMRSEC"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCAMR:AMRMIN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCAMR:AMRHOUR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCAMR:AMRDAY"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCAMR:AMRDATE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCAMR:AMRMON"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["RTC:RTCAMR:AMRYEAR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:DR:OE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "FIFOFull"},
	}}
	notDefault["UART:DR:BE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Error", Description: "indicates that the received data input was held LOW of longer than a full word transmission time(defined as start, data, parity and stop bits)"},
	}}
	notDefault["UART:DR:PE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Error", Description: "indicates that the parity of the received data character does not match the parity that the EPS and SPS bits in the line control register, UARTLCR_H select"},
	}}
	notDefault["UART:DR:FE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Error", Description: "indicates that the received character didn’t have a valid stop bit"},
	}}
	notDefault["UART:RSR:OE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "FIFOFull"},
	}}
	notDefault["UART:RSR:BE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Error", Description: "indicates that the received data input was held LOW of longer than a full word transmission time(defined as start, data, parity and stop bits)"},
	}}
	notDefault["UART:RSR:PE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Error", Description: "indicates that the parity of the received data character does not match the parity that the EPS and SPS bits in the line control register, UARTLCR_H select"},
	}}
	notDefault["UART:RSR:FE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Error", Description: "indicates that the received character didn’t have a valid stop bit"},
	}}
	notDefault["UART:ECR:OE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ECR:BE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ECR:PE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ECR:FE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:FR:RI"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "High"},
		{Value: "1", Name: "Low"},
	}}
	notDefault["UART:FR:TXFE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "TxHoldEmpty"},
		{Value: "1", Name: "TxFIFOEmpty"},
	}}
	notDefault["UART:FR:RXFF"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "RxHoldFull"},
		{Value: "1", Name: "RxFIFOFull"},
	}}
	notDefault["UART:FR:TXFF"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "TxHoldFull"},
		{Value: "1", Name: "TxFIFOFull"},
	}}
	notDefault["UART:FR:RXFE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "RxHoldEmpty"},
		{Value: "1", Name: "RxFIFOEmpty"},
	}}
	notDefault["UART:FR:BUSY"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Busy"},
	}}
	notDefault["UART:FR:DCD"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "DataCarrierDetect"},
	}}
	notDefault["UART:FR:DSR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "DataSetReady"},
	}}
	notDefault["UART:FR:CTS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "ClearToSend"},
	}}
	notDefault["UART:LCR_H:SPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:LCR_H:WLEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b00", Name: "5Bits"},
		{Value: "0b01", Name: "6Bits"},
		{Value: "0b10", Name: "7Bits"},
		{Value: "0b11", Name: "8Bits"},
	}}
	notDefault["UART:LCR_H:FEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:LCR_H:STP2"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "1Bit"},
		{Value: "1", Name: "2Bits"},
	}}
	notDefault["UART:LCR_H:EPS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Odd"},
		{Value: "1", Name: "Even"},
	}}
	notDefault["UART:LCR_H:PEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:LCR_H:BRK"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:CR:CTSEn"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:CR:RTSEn"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:CR:Out2"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "High"},
		{Value: "1", Name: "Low"},
	}}
	notDefault["UART:CR:Out1"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "High"},
		{Value: "1", Name: "Low"},
	}}
	notDefault["UART:CR:RTS"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "High"},
		{Value: "1", Name: "Low"},
	}}
	notDefault["UART:CR:DTR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "High"},
		{Value: "1", Name: "Low"},
	}}
	notDefault["UART:CR:RXE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:CR:TXE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:CR:SIRLP"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable", Description: "low level bits are transmitted as an active high pulse with a width of 3/16th of the bit period."},
		{Value: "1", Name: "Enable", Description: "low level bits are transmitted with a pulse width that is 3 times the period of the IrLPBaud16 input signal, regardless of the selected bit rate."},
	}}
	notDefault["UART:CR:SIREN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:CR:UARTEN"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IFLS:RXIFLSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Full1_8"},
		{Value: "0b001", Name: "Full1_4"},
		{Value: "0b010", Name: "Full1_2"},
		{Value: "0b011", Name: "Full3_4"},
		{Value: "0b100", Name: "Full7_8"},
	}}
	notDefault["UART:IFLS:TXIFLSEL"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0b000", Name: "Full1_8"},
		{Value: "0b001", Name: "Full1_4"},
		{Value: "0b010", Name: "Full1_2"},
		{Value: "0b011", Name: "Full3_4"},
		{Value: "0b100", Name: "Full7_8"},
	}}
	notDefault["UART:IMSC:OEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:BEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:PEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:FEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:RTIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:TXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:RXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:DSRMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:DCDMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:CTSMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:IMSC:RIMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART::OE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: ""},
	}}
	notDefault["UART:RIS:BEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:PEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:FEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:RTIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:TXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:RXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:DSRMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:DCDMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:CTSMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:RIS:RIMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:OEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:BEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:PEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:FEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:RTIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:TXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:RXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:DSRMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:DCDMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:CTSMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:MIS:RIMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Set"},
	}}
	notDefault["UART:ICR:OEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:BEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:PEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:FEIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:RTIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:TXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:RXIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:DSRMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:DCDMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:CTSMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:ICR:RIMIM"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "1", Name: "Clear"},
	}}
	notDefault["UART:DMACR:DMAONERR"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:DMACR:TXDMAE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	notDefault["UART:DMACR:RXDMAE"] = svd.EnumeratedValues{EnumeratedValue: []svd.EnumeratedValue{
		{Value: "0", Name: "Disable"},
		{Value: "1", Name: "Enable"},
	}}
	if v, ok := notDefault[typeDefName+":"+regName+":"+fieldName]; ok {
		return &v
	}
	return nil
}
