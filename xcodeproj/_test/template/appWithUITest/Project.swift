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
                        Target(name: "MyAppUITest",
                               platform: .iOS,
                               product: .uiTests,
                               bundleId: "io.bitrise.MyAppUITests",
                               infoPlist: .extendingDefault(with: [:]),
                               dependencies: [
                                    .sdk(name: "XCTest.framework", status: .required),
                                    .target(name: "MyApp")
                                ]),
                      ])