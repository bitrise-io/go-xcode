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
                                ]),
                        Target(name: "MyAppUniTest",
                               platform: .iOS,
                               product: .unitTests,
                               bundleId: "io.bitrise.MyAppUnitTests",
                               infoPlist: .extendingDefault(with: [:]),
                               dependencies: [
                                    .sdk(name: "XCTest", type: .framework, status: .required),
                                    .target(name: "MyApp")
                                ]),
                      ])