package testdata

// DeviceListXcode27 is a trimmed `xcrun simctl list --json` output of an Xcode 27 stack.
// The iOS 27.1 runtime supports only the iPhone Duo device type, and iOS 27.2 is installed twice:
// an iPhone Duo-only runtime (listed first) and a generic one, both reporting version 27.2.
// The iPhone 8 device type is not supported by any of the runtimes.
const DeviceListXcode27 = `
{
  "devicetypes" : [
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-18-Pro",
      "name" : "iPhone 18 Pro"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
      "name" : "iPhone Duo"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
      "name" : "iPhone 17"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro",
      "name" : "iPhone 16 Pro"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-13",
      "name" : "iPhone 13"
    },
    {
      "productFamily" : "iPad",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-13-inch-M5-12GB",
      "name" : "iPad Pro 13-inch (M5)"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
      "name" : "iPhone 8"
    }
  ],
  "runtimes" : [
    {
      "isAvailable" : true,
      "version" : "18.6",
      "buildversion" : "22G86",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone 16 Pro",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro"
        },
        {
          "name" : "iPhone 13",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-13"
        }
      ],
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-18-6",
      "platform" : "iOS",
      "name" : "iOS 18.6"
    },
    {
      "isAvailable" : true,
      "version" : "27.0",
      "buildversion" : "24A434",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone 18 Pro",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-18-Pro"
        },
        {
          "name" : "iPhone 17",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17"
        },
        {
          "name" : "iPhone 16 Pro",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro"
        },
        {
          "name" : "iPhone 13",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-13"
        },
        {
          "name" : "iPad Pro 13-inch (M5)",
          "productFamily" : "iPad",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-13-inch-M5-12GB"
        }
      ],
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-27-0",
      "platform" : "iOS",
      "name" : "iOS 27.0"
    },
    {
      "isAvailable" : true,
      "version" : "27.1",
      "buildversion" : "24A94232",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone Duo",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo"
        }
      ],
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-27-1",
      "platform" : "iOS",
      "name" : "iOS 27.1"
    },
    {
      "isAvailable" : true,
      "version" : "27.2",
      "buildversion" : "24B91011",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone Duo",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo"
        }
      ],
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4",
      "platform" : "iOS",
      "name" : "iOS 27.2"
    },
    {
      "isAvailable" : true,
      "version" : "27.2",
      "buildversion" : "24B5110a",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone 18 Pro",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-18-Pro"
        },
        {
          "name" : "iPhone 17",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17"
        },
        {
          "name" : "iPhone 16 Pro",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro"
        },
        {
          "name" : "iPhone 13",
          "productFamily" : "iPhone",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-13"
        },
        {
          "name" : "iPad Pro 13-inch (M5)",
          "productFamily" : "iPad",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-13-inch-M5-12GB"
        }
      ],
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-27-2",
      "platform" : "iOS",
      "name" : "iOS 27.2"
    }
  ],
  "devices" : {
    "com.apple.CoreSimulator.SimRuntime.iOS-27-0" : [
      {
        "udid" : "995D6B4B-41BD-43DE-9124-5735930D9A5A",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-18-Pro",
        "state" : "Shutdown",
        "name" : "iPhone 18 Pro"
      },
      {
        "udid" : "771C13EC-32E5-48E3-A58C-CA6C3D10B1DD",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
        "state" : "Shutdown",
        "name" : "iPhone 17"
      },
      {
        "udid" : "11DF5746-2EE4-439C-93EB-F57AD35271EB",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-13",
        "state" : "Shutdown",
        "name" : "Bitrise iOS default"
      },
      {
        "udid" : "3D24F467-BE8C-4D08-AF62-5AE97185E52D",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-13-inch-M5-12GB",
        "state" : "Shutdown",
        "name" : "iPad Pro 13-inch (M5)"
      }
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-27-1" : [
      {
        "udid" : "03F49682-6BB2-4AD7-A71D-28FA34A0FB31",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
        "state" : "Shutdown",
        "name" : "Bitrise iOS default"
      },
      {
        "udid" : "5BC7AC73-1349-4535-BDCF-28679722AA8B",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
        "state" : "Shutdown",
        "name" : "iPhone Duo"
      }
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-27-2" : [
      {
        "udid" : "96CF0E04-7A22-4A25-A43E-1C455DA678FF",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-18-Pro",
        "state" : "Shutdown",
        "name" : "iPhone 18 Pro"
      },
      {
        "udid" : "A54C9770-D487-455C-8BA2-9BDEFE5DD4E2",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-13",
        "state" : "Shutdown",
        "name" : "Bitrise iOS default"
      },
      {
        "udid" : "3C55B1E1-6DD4-4E2C-A8F4-7CB8ED7F9451",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPad-Pro-13-inch-M5-12GB",
        "state" : "Shutdown",
        "name" : "iPad Pro 13-inch (M5)"
      },
      {
        "udid" : "FB72DF49-C0D1-4F75-ADC3-D42850028AD6",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
        "state" : "Shutdown",
        "name" : "iPhone 17"
      }
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-18-6" : [
      {
        "udid" : "987C8F95-93D0-4017-8402-078E881203CD",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-16-Pro",
        "state" : "Shutdown",
        "name" : "iPhone 16 Pro"
      }
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4" : [
      {
        "udid" : "524E5A02-EC53-4129-BB01-E6832E4FE238",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
        "state" : "Shutdown",
        "name" : "iPhone Duo"
      }
    ]
  },
  "pairs" : {
  }
}
`
