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
                      ])