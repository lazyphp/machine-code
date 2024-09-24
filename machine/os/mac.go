/*
author: superl[N.S.T]
github: https://github.com/super-l/
desc: 获取mac操作系统的相关硬件基础编码信息
*/
package os

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lazyphp/machine-code/machine/types"
)

type MacMachine struct{}

var macMachineData types.Information

type macXmlStruct struct {
	XMLName xml.Name           `xml:"plist"`
	Array   macDataArrayStruct `xml:"array"`
}

type macDataArrayStruct struct {
	Dict macDictStruct `xml:"dict"` // 读取user数组
}

type macDictStruct struct {
	Key    []string           `xml:"key"`
	Real   []string           `xml:"real"`
	String []string           `xml:"string"`
	Array  macDictArrayStruct `xml:"array"`
}

type macDictArrayStruct struct {
	Dict macDictItemStruct `xml:"dict"` // 读取user数组
}

type macDictItemStruct struct {
	Key     []string `xml:"key"`
	Integer []int    `xml:"integer"`
	String  []string `xml:"string"`
}

func (mac MacMachine) GetMachine() types.Information {
	platformUUID, _ := mac.GetPlatformUUID()
	boardSerialNumber, _ := mac.GetBoardSerialNumber()
	cpuSerialNumber, _ := mac.GetCpuSerialNumber()

	machineData := types.Information{
		PlatformUUID:      platformUUID,
		BoardSerialNumber: boardSerialNumber,
		CpuSerialNumber:   cpuSerialNumber,
	}
	return machineData
}

func (mac MacMachine) GetBoardSerialNumber() (data string, err error) {
	sysInfo, err := mac.GetMacSysInfo()
	if err != nil {
		return "", err
	}
	return sysInfo.BoardSerialNumber, err
}

func (mac MacMachine) GetPlatformUUID() (UUID string, err error) {
	sysInfo, err := mac.GetMacSysInfo()
	if err != nil {
		return "", err
	}
	return sysInfo.PlatformUUID, err
}

func (mac MacMachine) GetCpuSerialNumber() (cpuId string, err error) {
	sysInfo, err := mac.GetMacSysInfo()
	if err != nil {
		return "", err
	}
	return sysInfo.CpuSerialNumber, err
}

func (mac MacMachine) GetMacSysInfo() (data types.Information, err error) {
	var cmd *exec.Cmd
	cmd = exec.Command("system_profiler", "SPHardwareDataType", "-xml")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = cmd.Wait()
	if err == nil {
		macMachineData, err = mac.macXmlToData(out.String())
		if err == nil {
			macMachineData.CpuSerialNumber, _ = mac.getCpuSerialNumberBase()
		}
		return macMachineData, nil
	} else {
		return types.Information{}, err
	}
}

func (MacMachine) macXmlToData(xmlcontent string) (types.Information, error) {
	x := macXmlStruct{}
	err := xml.Unmarshal([]byte(xmlcontent), &x)
	if err != nil {
		return types.Information{}, err
	} else {
		count := len(x.Array.Dict.Array.Dict.String)
		serialData := types.Information{
			PlatformUUID:      x.Array.Dict.Array.Dict.String[count-2],
			BoardSerialNumber: x.Array.Dict.Array.Dict.String[count-1],
			CpuSerialNumber:   "",
		}
		return serialData, nil
	}
}

func (mac MacMachine) getCpuSerialNumberBase() (cpuId string, err error) {
	// sysctl -x machdep.cpu.signature
	var cmd *exec.Cmd
	cmd = exec.Command("sysctl", "-x", "machdep.cpu.signature")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err == nil {
		cpuId = out.String()
		cpuId = strings.Replace(cpuId, " ", "", -1)
		cpuId = strings.Replace(cpuId, "\n", "", -1)
		cpuId = strings.Replace(cpuId, "machdep.cpu.signature:", "", -1)
		return cpuId, nil
	} else {
		return "", err
	}
}
