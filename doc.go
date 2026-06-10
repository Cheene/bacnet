// Copyright 2024 The BACnet Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

/*
Package bacnet provides a Go implementation of the BACnet/IP protocol stack
for building automation and control systems.

BACnet (Building Automation and Control Networks) is a data communication
protocol for building automation and control systems. It is designed to
facilitate communication between different devices and systems in buildings,
such as HVAC, lighting, security, and fire safety systems.

This package implements the following BACnet services:

Confirmed Services:
- ReadProperty (12)
- ReadPropertyMultiple (14)
- WriteProperty (15)
- WritePropertyMultiple (16)

Unconfirmed Services:
- IAm (0)
- WhoIs (8)

Key Features:
- Full BACnet/IP protocol support
- Device discovery (WhoIs/IAm)
- Object access (ReadProperty, WriteProperty)
- Transaction management with TSM (Transaction State Machine)
- Thread-safe design with connection pooling
- Support for multiple network interfaces

Data Collection Flow:
The typical data collection process involves the following steps:

1. Client Initialization: Create a BACnet client with appropriate network
   configuration (IP address, subnet, port).

2. Device Discovery: Use WhoIs service to discover all BACnet devices on the
   network. This sends a broadcast message to which all devices respond
   with their device ID and network address.

3. Object Discovery: Once devices are discovered, use Objects() method to
   retrieve all objects from a specific device. This includes Analog Inputs,
   Binary Inputs, Analog Outputs, Binary Outputs, and other object types.

4. Data Reading: Use ReadProperty() to read individual property values or
   ReadMultiProperty() to read multiple properties in a single request.

5. Data Writing: Use WriteProperty() or WriteMultiProperty() to write values
   to device objects.

6. Cleanup: Always close the client when done to release resources.

Example Usage:
	// Create client
	client, err := bacnet.NewClient(&bacnet.ClientBuilder{
	    Ip:         "192.168.1.100",
	    SubnetCIDR: 24,
	    Port:       47808,
	})
	if err != nil {
	    log.Fatal(err)
	}
	defer client.Close()

	// Start message loop
	go client.ClientRun()

	// Discover devices
	devices, err := client.WhoIs(&bacnet.WhoIsOpts{
	    Low:  0,
	    High: 4194304,
	})

	// Read property from device
	result, err := client.ReadProperty(device, btypes.PropertyData{
	    Object: btypes.Object{
	        ID: btypes.ObjectID{
	            Type:     btypes.AnalogInput,
	            Instance: 1,
	        },
	        Properties: []btypes.Property{
	            {Type: btypes.PropPresentValue},
	        },
	    },
	})

For more detailed information, see the README.md file.

中文说明：
BACnet（楼宇自动化和控制网络）是用于楼宇自动化和控制系统的数据通信协议。
它旨在促进楼宇中不同设备和系统之间的通信，如暖通空调（HVAC）、照明、安全和消防系统。

数据采集流程：
典型的数据采集过程包括以下步骤：

1. 客户端初始化：使用适当的网络配置创建BACnet客户端（IP地址、子网、端口）。

2. 设备发现：使用WhoIs服务发现网络上的所有BACnet设备。这会发送广播消息，
   所有设备都会响应其设备ID和网络地址。

3. 对象发现：发现设备后，使用Objects()方法从特定设备检索所有对象。
   包括模拟输入、二进制输入、模拟输出、二进制输出等对象类型。

4. 数据读取：使用ReadProperty()读取单个属性值，或使用ReadMultiProperty()
   在单个请求中读取多个属性。

5. 数据写入：使用WriteProperty()或WriteMultiProperty()向设备对象写入值。

6. 清理：完成操作后始终关闭客户端以释放资源。

参考文档：
- ANSI/ASHRAE Standard 135-2020
- http://www.bacnet.org/
*/
package bacnet