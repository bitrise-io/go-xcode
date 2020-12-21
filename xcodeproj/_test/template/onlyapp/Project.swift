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
                      ])