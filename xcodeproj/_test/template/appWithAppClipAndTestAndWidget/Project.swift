import ProjectDescription

let project = Project(name: "MyApp",
                      organizationName: "Bitrise",
                      targets: [
                        Target(name: "MyApp",
                               platform: .iOS,
                               product: .app,
                               bundleId: "io.bitrise.MyApp",
                               infoPlist: .extendingDefault(with: [:]),
                               dependencies: [
                                 .target(name: "MyAppClip"),
                                 .target(name: "MyAppWidget"),
                               ]),
                        Target(name: "MyAppClip",
                               platform: .iOS,
                               product: .appClip,
                               bundleId: "io.bitrise.MyApp.Clip",
                               infoPlist: .extendingDefault(with: [:]),
                               entitlements: "AppClip.entitlements",
                               dependencies: [
                                    .sdk(name: "AppClip.framework", status: .required),
                                ]),
                        Target(name: "MyAppUniTest",
                               platform: .iOS,
                               product: .unitTests,
                               bundleId: "io.bitrise.MyAppUnitTests",
                               infoPlist: .extendingDefault(with: [:]),
                               dependencies: [
                                    .sdk(name: "XCTest.framework", status: .required),
                                    .target(name: "MyApp")
                                ]),
                        Target(name: "MyAppWidget",
                               platform: .iOS,
                               product: .appExtension,
                               bundleId: "io.bitrise.MyAppWidget",
                               infoPlist: .extendingDefault(with: [:]),
                               dependencies: [
                                ]),
                      ])