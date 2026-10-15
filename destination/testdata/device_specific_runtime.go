package testdata

// DeviceListWithDeviceSpecificRuntime mimics Xcode 27.2 Beta 3, which installs a generic iOS 27.2 runtime and
// a device-specific one (for iPhone Duo) with the same version. The device-specific runtime is listed first.
const DeviceListWithDeviceSpecificRuntime = `
{
  "devicetypes" : [
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-12-Pro-Max",
      "name" : "iPhone 12 Pro Max"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
      "name" : "iPhone 17"
    },
    {
      "productFamily" : "iPhone",
      "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
      "name" : "iPhone Duo"
    }
  ],
  "runtimes" : [
    {
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4",
      "platform" : "iOS",
      "version" : "27.2",
      "isAvailable" : true,
      "name" : "iOS 27.2",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone Duo",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
          "productFamily" : "iPhone"
        }
      ]
    },
    {
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-27-2",
      "platform" : "iOS",
      "version" : "27.2",
      "isAvailable" : true,
      "name" : "iOS 27.2",
      "supportedDeviceTypes" : [
        {
          "name" : "iPhone 12 Pro Max",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-12-Pro-Max",
          "productFamily" : "iPhone"
        },
        {
          "name" : "iPhone 17",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
          "productFamily" : "iPhone"
        }
      ]
    }
  ],
  "devices" : {
    "com.apple.CoreSimulator.SimRuntime.iOS-27-2-iPhone19-4" : [
      {
        "udid" : "6F0E6D53-7F1B-4F0B-9C2E-1D2A4B3C5D01",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-Duo",
        "state" : "Shutdown",
        "name" : "iPhone Duo"
      }
    ],
    "com.apple.CoreSimulator.SimRuntime.iOS-27-2" : [
      {
        "udid" : "FB72DF49-C0D1-4F75-ADC3-D42850028AD6",
        "isAvailable" : true,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-17",
        "state" : "Shutdown",
        "name" : "iPhone 17"
      }
    ]
  }
}
`
