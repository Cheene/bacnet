package bacnet

import (
	"time"

	"github.com/anviod/bacnet/btypes"
	"github.com/anviod/bacnet/encoding"
)

// WhoIsOpts contains options for the WhoIs device discovery request.
//
// 中文说明：WhoIsOpts 包含 WhoIs 设备发现请求的选项。
type WhoIsOpts struct {
	Low             int             `json:"low"`              // Lower bound of device ID range (0 to 4194304)
	High            int             `json:"high"`             // Upper bound of device ID range
	GlobalBroadcast bool            `json:"global_broadcast"` // Use global broadcast (0xFFFF)
	NetworkNumber   uint16          `json:"network_number"`   // Target network number
	Destination     *btypes.Address `json:"-"`                // Specific destination address (optional)
}

// WhoIs discovers all BACnet devices on the network within the specified device ID range.
// It sends a broadcast WhoIs request and collects IAm responses from devices.
// The response includes device information such as Device ID, IP address, port,
// MaxAPDU size, segmentation support, and vendor ID.
//
// Parameters:
//
//	wh - WhoIsOpts containing the device ID range and other options
//
// Returns:
//
//	A list of discovered devices and any error encountered.
//
// Usage Notes:
//   - Use Low=0 and High=4194304 to discover all devices
//   - Specify a narrow range for targeted discovery to reduce network traffic
//   - Use Destination for unicast WhoIs requests
//   - The function handles duplicate device responses automatically
//
// 中文说明：WhoIs 发现网络上指定设备ID范围内的所有BACnet设备。
// 发送广播 WhoIs 请求并收集设备的 IAm 响应。
// 响应包括设备信息，如设备ID、IP地址、端口、MaxAPDU大小、分段支持和供应商ID。
//
// 参数：
//
//	wh - 包含设备ID范围和其他选项的 WhoIsOpts
//
// 返回：
//
//	发现的设备列表和遇到的任何错误。
//
// 使用注意：
//   - 使用 Low=0 和 High=4194304 发现所有设备
//   - 指定窄范围进行目标发现以减少网络流量
//   - 使用 Destination 进行单播 WhoIs 请求
//   - 函数自动处理重复设备响应
func (c *client) WhoIs(wh *WhoIsOpts) ([]btypes.Device, error) {
	dest := *c.dataLink.GetBroadcastAddress()
	if wh.Destination != nil {
		dest = *wh.Destination
	}

	enc := encoding.NewEncoder()
	low := wh.Low
	high := wh.High
	if wh.GlobalBroadcast {
		wh.NetworkNumber = btypes.GlobalBroadcast
	}
	if low <= 0 {
		low = 0
	}
	if high <= 0 {
		high = 4194304
	}

	dest.Net = wh.NetworkNumber
	npdu := &btypes.NPDU{
		Version:               btypes.ProtocolVersion,
		Destination:           &dest,
		Source:                c.dataLink.GetMyAddress(),
		IsNetworkLayerMessage: false,
		ExpectingReply:        false,
		Priority:              btypes.Normal,
		HopCount:              btypes.DefaultHopCount,
	}
	enc.NPDU(npdu)
	err := enc.WhoIs(int32(low), int32(high))
	if err != nil {
		return nil, err
	}

	var start, end int
	if low == -1 || high == -1 {
		start = 0
		end = maxInt
	} else {
		start = low
		end = high
	}

	errChan := make(chan error)
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, err = c.Send(dest, npdu, enc.Bytes(), nil)
		errChan <- err
	}()
	values, err := c.utsm.Subscribe(start, end)
	if err != nil {
		return nil, err
	}
	err = <-errChan
	if err != nil {
		return nil, err
	}

	uniqueMap := make(map[btypes.ObjectInstance]btypes.Device)
	var uniqueList []btypes.Device

	for _, v := range values {
		iam, ok := v.(btypes.IAm)
		if !ok {
			continue
		}
		if _, ok := uniqueMap[iam.ID.Instance]; ok {
			continue
		}

		dev := btypes.Device{
			DeviceID:     int(iam.ID.Instance),
			Addr:         iam.Addr,
			ID:           iam.ID,
			MaxApdu:      iam.MaxApdu,
			Segmentation: iam.Segmentation,
			Vendor:       iam.Vendor,
		}

		// Some BACnet/IP stacks send I-Am from an ephemeral UDP source port while
		// still accepting confirmed services on the configured BACnet/IP port.
		// When the caller used a unicast Who-Is destination, keep that address as
		// the confirmed-service destination instead of blindly copying the I-Am
		// packet source address.
		if wh.Destination != nil && !wh.Destination.IsBroadcast() {
			dev.Addr = *wh.Destination
		}
		if udpAddr, err := dev.Addr.UDPAddr(); err == nil {
			dev.Ip = udpAddr.IP.String()
			dev.Port = udpAddr.Port
		}

		uniqueMap[iam.ID.Instance] = dev
		uniqueList = append(uniqueList, dev)
	}
	return uniqueList, err
}
