package xcscheme

const schemeWithTestPlanContent = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme
   LastUpgradeVersion = "1300"
   version = "1.7">
   <BuildAction
      parallelizeBuildables = "YES"
      buildImplicitDependencies = "YES">
      <BuildActionEntries>
         <BuildActionEntry
            buildForTesting = "YES"
            buildForRunning = "YES"
            buildForProfiling = "YES"
            buildForArchiving = "YES"
            buildForAnalyzing = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "D2A5F1FF1F4A9144005CD714"
               BuildableName = "BullsEye.app"
               BlueprintName = "BullsEye"
               ReferencedContainer = "container:BullsEye.xcodeproj">
            </BuildableReference>
         </BuildActionEntry>
      </BuildActionEntries>
   </BuildAction>
   <TestAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      shouldUseLaunchSchemeArgsEnv = "YES">
      <TestPlans>
         <TestPlanReference
            reference = "container:FullTests.xctestplan"
            default = "YES">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:UnitTests.xctestplan">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:UITests.xctestplan">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:ParallelUITests.xctestplan">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:FailingTests.xctestplan">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:EventuallyFailingTests.xctestplan">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:EventuallySucceedingTests.xctestplan">
         </TestPlanReference>
         <TestPlanReference
            reference = "container:EventuallyFailingInMemoryTests.xctestplan">
         </TestPlanReference>
      </TestPlans>
      <Testables>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "13FB7CED267726620084066F"
               BuildableName = "BullsEyeTests.xctest"
               BlueprintName = "BullsEyeTests"
               ReferencedContainer = "container:BullsEye.xcodeproj">
            </BuildableReference>
         </TestableReference>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "13FB7CFA2677288A0084066F"
               BuildableName = "BullsEyeSlowTests.xctest"
               BlueprintName = "BullsEyeSlowTests"
               ReferencedContainer = "container:BullsEye.xcodeproj">
            </BuildableReference>
         </TestableReference>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "13FB7D0B26773C570084066F"
               BuildableName = "BullsEyeUITests.xctest"
               BlueprintName = "BullsEyeUITests"
               ReferencedContainer = "container:BullsEye.xcodeproj">
            </BuildableReference>
         </TestableReference>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "139E88A6268DBCDA0007755C"
               BuildableName = "BullsEyeFailingTests.xctest"
               BlueprintName = "BullsEyeFailingTests"
               ReferencedContainer = "container:BullsEye.xcodeproj">
            </BuildableReference>
         </TestableReference>
      </Testables>
   </TestAction>
   <LaunchAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      launchStyle = "0"
      useCustomWorkingDirectory = "NO"
      ignoresPersistentStateOnLaunch = "NO"
      debugDocumentVersioning = "YES"
      debugServiceExtension = "internal"
      allowLocationSimulation = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "D2A5F1FF1F4A9144005CD714"
            BuildableName = "BullsEye.app"
            BlueprintName = "BullsEye"
            ReferencedContainer = "container:BullsEye.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </LaunchAction>
   <ProfileAction
      buildConfiguration = "Release"
      shouldUseLaunchSchemeArgsEnv = "YES"
      savedToolIdentifier = ""
      useCustomWorkingDirectory = "NO"
      debugDocumentVersioning = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "D2A5F1FF1F4A9144005CD714"
            BuildableName = "BullsEye.app"
            BlueprintName = "BullsEye"
            ReferencedContainer = "container:BullsEye.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </ProfileAction>
   <AnalyzeAction
      buildConfiguration = "Debug">
   </AnalyzeAction>
   <ArchiveAction
      buildConfiguration = "Release"
      revealArchiveInOrganizer = "YES">
   </ArchiveAction>
</Scheme>
`

const schemeContent = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme
   LastUpgradeVersion = "0800"
   version = "1.3">
   <BuildAction
      parallelizeBuildables = "YES"
      buildImplicitDependencies = "YES">
      <BuildActionEntries>
         <BuildActionEntry
            buildForTesting = "YES"
            buildForRunning = "YES"
            buildForProfiling = "YES"
            buildForArchiving = "YES"
            buildForAnalyzing = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "BA3CBE7419F7A93800CED4D5"
               BuildableName = "ios-simple-objc.app"
               BlueprintName = "ios-simple-objc"
               ReferencedContainer = "container:ios-simple-objc.xcodeproj">
            </BuildableReference>
         </BuildActionEntry>
         <BuildActionEntry
            buildForTesting = "YES"
            buildForRunning = "YES"
            buildForProfiling = "NO"
            buildForArchiving = "NO"
            buildForAnalyzing = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "BA3CBE9019F7A93900CED4D5"
               BuildableName = "ios-simple-objcTests.xctest"
               BlueprintName = "ios-simple-objcTests"
               ReferencedContainer = "container:ios-simple-objc.xcodeproj">
            </BuildableReference>
         </BuildActionEntry>
      </BuildActionEntries>
   </BuildAction>
   <TestAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      shouldUseLaunchSchemeArgsEnv = "YES">
      <Testables>
         <TestableReference
            skipped = "NO">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "BA3CBE9019F7A93900CED4D5"
               BuildableName = "ios-simple-objcTests.xctest"
               BlueprintName = "ios-simple-objcTests"
               ReferencedContainer = "container:ios-simple-objc.xcodeproj">
            </BuildableReference>
         </TestableReference>
         <TestableReference
            skipped = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "BA4CBE9019F7A93900CED4D5"
               BuildableName = "ios-simple-objcTests2.xctest"
               BlueprintName = "ios-simple-objcTests2"
               ReferencedContainer = "container:ios-simple-objc.xcodeproj">
            </BuildableReference>
         </TestableReference>
      </Testables>
      <MacroExpansion>
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "BA3CBE7419F7A93800CED4D5"
            BuildableName = "ios-simple-objc.app"
            BlueprintName = "ios-simple-objc"
            ReferencedContainer = "container:ios-simple-objc.xcodeproj">
         </BuildableReference>
      </MacroExpansion>
      <AdditionalOptions>
      </AdditionalOptions>
   </TestAction>
   <LaunchAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      launchStyle = "0"
      useCustomWorkingDirectory = "NO"
      ignoresPersistentStateOnLaunch = "NO"
      debugDocumentVersioning = "YES"
      debugServiceExtension = "internal"
      allowLocationSimulation = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "BA3CBE7419F7A93800CED4D5"
            BuildableName = "ios-simple-objc.app"
            BlueprintName = "ios-simple-objc"
            ReferencedContainer = "container:ios-simple-objc.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
      <AdditionalOptions>
      </AdditionalOptions>
   </LaunchAction>
   <ProfileAction
      buildConfiguration = "Release"
      shouldUseLaunchSchemeArgsEnv = "YES"
      savedToolIdentifier = ""
      useCustomWorkingDirectory = "NO"
      debugDocumentVersioning = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "BA3CBE7419F7A93800CED4D5"
            BuildableName = "ios-simple-objc.app"
            BlueprintName = "ios-simple-objc"
            ReferencedContainer = "container:ios-simple-objc.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </ProfileAction>
   <AnalyzeAction
      buildConfiguration = "Debug">
   </AnalyzeAction>
   <ArchiveAction
      buildConfiguration = "Release"
      revealArchiveInOrganizer = "YES">
   </ArchiveAction>
</Scheme>
`
