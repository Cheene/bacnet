## Documentation

- [English Documentation](README.md) (This file)
- [中文文档](README_CN.md)
  
# BACnet Protocol Stack

A Go implementation of the BACnet/IP protocol stack for building automation and control systems.

## Features

- **BACnet/IP Protocol**: Full support for BACnet/IP communication
- **Device Discovery**: Who-Is and I-Am services for network device discovery
- **Object Access**: ReadProperty, ReadMultipleProperty, WriteProperty, WriteMultipleProperty
- **Network Management**: What-Is-Network-Number, Who-Is-Router-To-Network
- **Transaction Management**: TSM (Transaction State Machine) for confirmed services
- **Concurrency**: Thread-safe design with connection pooling

## Installation

```bash
go get github.com/anviod/bacnet
```

## Quick Start

### Basic Device Discovery

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/anviod/bacnet"
    "github.com/anviod/bacnet/btypes"
)

func main() {
    // Create a BACnet client
    client, err := bacnet.NewClient(&bacnet.ClientBuilder{
        Ip:         "192.168.1.100",
        SubnetCIDR: 24,
        Port:       47808, // Default BACnet port (0xBAC0)
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Start the client message loop
    go client.ClientRun()

    // Discover all devices on the network
    devices, err := client.WhoIs(&bacnet.WhoIsOpts{
        Low:  0,
        High: 4194304, // Max BACnet device ID
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Discovered %d devices\n", len(devices))
    for _, dev := range devices {
        fmt.Printf("Device ID: %d, IP: %s:%d\n", dev.DeviceID, dev.Ip, dev.Port)
    }
}
```

---

## Data Collection Flow (采集流程)

The BACnet data collection process consists of six key steps:

### Step 1: Client Initialization

Before any communication can occur, a BACnet client must be created with appropriate network configuration.

```go
client, err := bacnet.NewClient(&bacnet.ClientBuilder{
    Ip:         "192.168.1.100",  // Local IP address
    SubnetCIDR: 24,                // Subnet mask (e.g., /24)
    Port:       47808,             // BACnet port (default: 47808)
})
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

**Configuration Options:**
- `Ip`: Local IP address to bind to
- `Interface`: Network interface name (alternative to Ip)
- `SubnetCIDR`: Subnet CIDR notation (e.g., 24 for 255.255.255.0)
- `Port`: BACnet UDP port (default: 47808 = 0xBAC0)
- `MaxPDU`: Maximum PDU size (default: 1476)

### Step 2: Start Message Loop

The client message loop must be started in a goroutine to handle incoming messages:

```go
go client.ClientRun()
```

**Important Notes:**
- Must be called before making any requests
- Runs continuously until the client is closed
- Handles message decoding and routing

### Step 3: Device Discovery (WhoIs)

Discover BACnet devices on the network using the WhoIs service:

```go
devices, err := client.WhoIs(&bacnet.WhoIsOpts{
    Low:  0,             // Device ID lower bound
    High: 4194304,       // Device ID upper bound (max)
})
```

**Discovery Options:**
- `Low`: Lower bound of device ID range (0 to 4194304)
- `High`: Upper bound of device ID range
- `GlobalBroadcast`: Use global broadcast address (0xFFFF)
- `Destination`: Specific target address for unicast discovery

**Best Practices:**
- Use narrow ID ranges for targeted discovery to reduce network traffic
- Avoid using full range (0-4194304) on large networks
- Cache discovered devices to avoid repeated discovery

### Step 4: Object Discovery

Retrieve all objects from a discovered device:

```go
scannedDevice, err := client.Objects(devices[0])
if err != nil {
    log.Printf("Failed to scan objects: %v", err)
    return
}

// Access specific object types
aiObjects := scannedDevice.Objects[btypes.AnalogInput]
biObjects := scannedDevice.Objects[btypes.BinaryInput]
aoObjects := scannedDevice.Objects[btypes.AnalogOutput]
boObjects := scannedDevice.Objects[btypes.BinaryOutput]
```

**Supported Object Types:**
- `AnalogInput` (0): Analog input points (e.g., temperature sensors)
- `AnalogOutput` (1): Analog output points (e.g., valves, dampers)
- `AnalogValue` (2): Analog value objects
- `BinaryInput` (3): Binary input points (e.g., contact sensors)
- `BinaryOutput` (4): Binary output points (e.g., relays)
- `BinaryValue` (5): Binary value objects
- `Device` (8): BACnet device objects
- `MultiStateInput` (13): Multi-state input points
- `MultiStateOutput` (14): Multi-state output points
- `TrendLog` (20): Trend log objects

### Step 5: Data Reading

Read property values from device objects.

#### Read Single Property

```go
result, err := client.ReadProperty(device, btypes.PropertyData{
    Object: btypes.Object{
        ID: btypes.ObjectID{
            Type:     btypes.AnalogInput,
            Instance: 1,
        },
        Properties: []btypes.Property{
            {
                Type:       btypes.PropPresentValue,
                ArrayIndex: btypes.ArrayAll,
            },
        },
    },
})
```

#### Read Multiple Properties (Batch)

For better performance, use ReadMultiProperty to read multiple properties in one request:

```go
result, err := client.ReadMultiProperty(device, btypes.MultiplePropertyData{
    Objects: []btypes.Object{
        {
            ID: btypes.ObjectID{Type: btypes.AnalogInput, Instance: 1},
            Properties: []btypes.Property{
                {Type: btypes.PropPresentValue},
                {Type: btypes.PropUnits},
                {Type: btypes.PropDescription},
            },
        },
        {
            ID: btypes.ObjectID{Type: btypes.AnalogInput, Instance: 2},
            Properties: []btypes.Property{
                {Type: btypes.PropPresentValue},
            },
        },
    },
})
```

**Common Properties:**
- `PropPresentValue` (85): Current value of the object
- `PropUnits` (117): Engineering units
- `PropDescription` (28): Object description
- `PropObjectName` (77): Object name
- `PropObjectType` (79): Object type
- `PropObjectIdentifier` (75): Object identifier
- `PropObjectList` (76): List of objects in device

### Step 6: Data Writing

Write values to device objects.

```go
err := client.WriteProperty(device, btypes.PropertyData{
    Object: btypes.Object{
        ID: btypes.ObjectID{
            Type:     btypes.AnalogOutput,
            Instance: 1,
        },
        Properties: []btypes.Property{
            {
                Type:       btypes.PropPresentValue,
                ArrayIndex: btypes.ArrayAll,
                Data:       float64(25.5),
                Priority:   btypes.Normal,
            },
        },
    },
})
```

**Write Priority Levels:**
- `LifeSafety` (3): Life safety operations
- `CriticalEquipment` (2): Critical equipment control
- `Urgent` (1): Urgent operations
- `Normal` (0): Normal operations

---

## Advanced Usage

### Complete Integration Flow

```go
func completeIntegration(client bacnet.Client) error {
    // Step 1: Discover devices
    devices, err := client.WhoIs(&bacnet.WhoIsOpts{
        Low:  0,
        High: 4194304,
    })
    if err != nil {
        return fmt.Errorf("whois failed: %v", err)
    }
    if len(devices) == 0 {
        return fmt.Errorf("no devices found")
    }

    device := devices[0]
    fmt.Printf("Found device: ID=%d, IP=%s:%d\n", device.DeviceID, device.Ip, device.Port)

    // Step 2: Scan objects
    scannedDevice, err := client.Objects(device)
    if err != nil {
        return fmt.Errorf("object scan failed: %v", err)
    }

    // Step 3: Find target point
    aiObjects := scannedDevice.Objects[btypes.AnalogInput]
    targetPoint, ok := aiObjects[1]
    if !ok {
        return fmt.Errorf("target point not found")
    }
    fmt.Printf("Found target point: %s\n", targetPoint.Name)

    // Step 4: Read present value
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
    if err != nil {
        return fmt.Errorf("read property failed: %v", err)
    }
    fmt.Printf("Present Value: %v\n", result.Object.Properties[0].Data)

    // Step 5: Write to AnalogValue
    writeErr := client.WriteProperty(device, btypes.PropertyData{
        Object: btypes.Object{
            ID: btypes.ObjectID{
                Type:     btypes.AnalogValue,
                Instance: 1,
            },
            Properties: []btypes.Property{
                {
                    Type:       btypes.PropPresentValue,
                    ArrayIndex: btypes.ArrayAll,
                    Data:       float64(25.5),
                    Priority:   btypes.Normal,
                },
            },
        },
    })
    if writeErr != nil {
        return fmt.Errorf("write property failed: %v", writeErr)
    }
    fmt.Println("Write successful")

    return nil
}
```

### Read with Timeout

Use timeout variants for better control over request timing:

```go
result, err := client.ReadPropertyWithTimeout(device, propertyData, 5*time.Second)
```

### Error Handling Patterns

```go
func safeReadProperty(client bacnet.Client, device btypes.Device, objID btypes.ObjectID) (interface{}, error) {
    result, err := client.ReadProperty(device, btypes.PropertyData{
        Object: btypes.Object{
            ID: objID,
            Properties: []btypes.Property{
                {Type: btypes.PropPresentValue},
            },
        },
    })
    
    if err != nil {
        // Handle specific error types
        if strings.Contains(err.Error(), "timeout") {
            return nil, fmt.Errorf("device %d did not respond", device.DeviceID)
        }
        if strings.Contains(err.Error(), "no such object") {
            return nil, fmt.Errorf("object %s not found", objID.Type)
        }
        return nil, err
    }
    
    if len(result.Object.Properties) == 0 {
        return nil, fmt.Errorf("no properties returned")
    }
    
    return result.Object.Properties[0].Data, nil
}
```

---

## API Reference

### Client Interface

```go
type Client interface {
    io.Closer
    IsRunning() bool
    ClientRun()
    
    // Device Discovery
    WhoIs(wh *WhoIsOpts) ([]btypes.Device, error)
    IAm(dest btypes.Address, iam btypes.IAm) error
    
    // Network Management
    WhatIsNetworkNumber() []*btypes.Address
    WhoIsRouterToNetwork() (resp *[]btypes.Address)
    
    // Object Access
    Objects(dev btypes.Device) (btypes.Device, error)
    ReadProperty(dest btypes.Device, rp btypes.PropertyData) (btypes.PropertyData, error)
    ReadMultiProperty(dev btypes.Device, rp btypes.MultiplePropertyData) (btypes.MultiplePropertyData, error)
    WriteProperty(dest btypes.Device, wp btypes.PropertyData) error
    WriteMultiProperty(dev btypes.Device, wp btypes.MultiplePropertyData) error
    
    // Timeout variants
    ReadPropertyWithTimeout(dest btypes.Device, rp btypes.PropertyData, timeout time.Duration) (btypes.PropertyData, error)
    ReadMultiPropertyWithTimeout(dev btypes.Device, rp btypes.MultiplePropertyData, timeout time.Duration) (btypes.MultiplePropertyData, error)
}
```

### WhoIs Options

```go
type WhoIsOpts struct {
    Low             int             // Device ID lower bound (0 to 4194304)
    High            int             // Device ID upper bound
    GlobalBroadcast bool            // Use global broadcast (0xFFFF)
    NetworkNumber   uint16          // Target network number
    Destination     *btypes.Address // Specific destination (optional)
}
```

---

## Configuration

### ClientBuilder Options

```go
type ClientBuilder struct {
    DataLink   datalink.DataLink // Custom data link (optional)
    Interface  string            // Network interface name (e.g., "eth0")
    Ip         string            // IP address
    Port       int               // BACnet port (default: 47808)
    SubnetCIDR int               // Subnet CIDR (e.g., 24 for /24)
    MaxPDU     uint16            // Maximum PDU size (default: 1476)
}
```

### Constants

```go
// Protocol
const DefaultPort = 0xBAC0 // 47808
const MaxAPDU = 1476

// Network
const GlobalBroadcast = 0xFFFF
const DefaultHopCount = 255

// Priorities
const (
    LifeSafety        = 3
    CriticalEquipment = 2
    Urgent            = 1
    Normal            = 0
)
```

---

## Best Practices & Recommendations

### Network Considerations

1. **Port Binding**:
   - Default BACnet port is 47808 (0xBAC0)
   - Use different ports for testing to avoid conflicts
   - Bind to `0.0.0.0` to listen on all interfaces

2. **IP Address Binding**:
   - Avoid binding to the target device's IP address
   - For multi-subnet environments, configure subnet CIDR properly

3. **Broadcast Behavior**:
   - WhoIs uses broadcast by default
   - Use `Destination` for unicast requests
   - Broadcast may not work across VLANs or subnets

### Performance Optimization

1. **Batch Operations**:
   - Use `ReadMultiProperty` for reading multiple properties
   - Reduce network round-trips
   - Limit batch size based on device's MaxAPDU setting

2. **Concurrency**:
   - Client is thread-safe for concurrent operations
   - TSM limits concurrent confirmed transactions (default: 10)
   - Consider rate limiting for high-frequency operations

3. **Memory Management**:
   - Use buffer pool for efficient memory usage
   - Release resources promptly with `client.Close()`

### Error Handling

1. **Timeout Handling**:
   - Use `ReadPropertyWithTimeout` for explicit timeout control
   - Confirmed services include retry logic with exponential backoff
   - Implement application-level retry for critical operations

2. **Common Errors**:
   - `timeout`: Device did not respond within timeout
   - `invalid argument`: Invalid object type or property ID
   - `no such object`: Requested object does not exist
   - `access denied`: Insufficient permissions for write operations

---

## Common Issues & Troubleshooting

### Issue 1: No Devices Discovered

**Possible Causes:**
- Incorrect IP address or subnet configuration
- Firewall blocking BACnet port (47808)
- Devices on different VLAN/subnet
- Client not running (`ClientRun()` not called)

**Solutions:**
- Verify network configuration
- Check firewall rules
- Use Wireshark to monitor BACnet traffic
- Ensure `ClientRun()` is called before `WhoIs()`

### Issue 2: ReadProperty Fails with Timeout

**Possible Causes:**
- Device not responding
- Incorrect device address
- Network connectivity issues
- Device busy or overloaded

**Solutions:**
- Verify device is reachable via ping
- Check device address (some devices use different ports for confirmed services)
- Increase timeout value
- Implement retry logic

### Issue 3: WriteProperty Returns "Access Denied"

**Possible Causes:**
- Insufficient permissions
- Write protection enabled on device
- Incorrect priority level

**Solutions:**
- Check device configuration for write permissions
- Verify priority level (use appropriate priority)
- Contact device manufacturer for access rights

### Issue 4: High Network Traffic

**Possible Causes:**
- Frequent WhoIs requests with full ID range
- Large batch operations exceeding MTU
- Broadcast storms

**Solutions:**
- Use targeted WhoIs with narrow ID ranges
- Limit batch size to stay within MaxAPDU
- Implement device discovery caching

---

## Testing

```bash
# Run all tests
go test ./...

# Run specific test
go test -v ./network/...

# Run acceptance tests
go test -v -run Acceptance

# Run real device integration test
go test . -run TestRealDeviceAcceptanceFlow -count=1 -v
```

## License

MIT License

## References

- [ANSI/ASHRAE Standard 135-2020](https://www.ashrae.org/standards-research/standards/ashrae-standard-135)
- [BACnet Protocol Specification](http://www.bacnet.org/)