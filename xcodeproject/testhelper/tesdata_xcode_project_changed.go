package testhelper

// XcodeProjectTestChanged project.pbxproj from https://github.com/bitrise-io/xcode-project-test, with codesign settings applied
const XcodeProjectTestChanged = `// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 50;
	objects = {

/* Begin PBXBuildFile section */
		7D0342F420F4BA280050B6A6 /* XcodeProjUITests.swift in Sources */ = {isa = PBXBuildFile; fileRef = 7D0342F320F4BA280050B6A6 /* XcodeProjUITests.swift */; };
		7D03431020F4BB070050B6A6 /* NotificationCenter.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = 7D03430F20F4BB070050B6A6 /* NotificationCenter.framework */; };
		7D03431320F4BB070050B6A6 /* TodayViewController.swift in Sources */ = {isa = PBXBuildFile; fileRef = 7D03431220F4BB070050B6A6 /* TodayViewController.swift */; };
		7D03431620F4BB070050B6A6 /* MainInterface.storyboard in Resources */ = {isa = PBXBuildFile; fileRef = 7D03431420F4BB070050B6A6 /* MainInterface.storyboard */; };
		7D03431A20F4BB070050B6A6 /* TodayExtension.appex in Embed App Extensions */ = {isa = PBXBuildFile; fileRef = 7D03430D20F4BB070050B6A6 /* TodayExtension.appex */; settings = {ATTRIBUTES = (RemoveHeadersOnCopy, ); }; };
		7D03432120F4BB8D0050B6A6 /* CloudKit.framework in Frameworks */ = {isa = PBXBuildFile; fileRef = 7D03432020F4BB8D0050B6A6 /* CloudKit.framework */; };
		7D5B360020E28EE80022BAE6 /* AppDelegate.swift in Sources */ = {isa = PBXBuildFile; fileRef = 7D5B35FF20E28EE80022BAE6 /* AppDelegate.swift */; };
		7D5B360220E28EE80022BAE6 /* ViewController.swift in Sources */ = {isa = PBXBuildFile; fileRef = 7D5B360120E28EE80022BAE6 /* ViewController.swift */; };
		7D5B360520E28EE80022BAE6 /* Main.storyboard in Resources */ = {isa = PBXBuildFile; fileRef = 7D5B360320E28EE80022BAE6 /* Main.storyboard */; };
		7D5B360720E28EEA0022BAE6 /* Assets.xcassets in Resources */ = {isa = PBXBuildFile; fileRef = 7D5B360620E28EEA0022BAE6 /* Assets.xcassets */; };
		7D5B360A20E28EEA0022BAE6 /* LaunchScreen.storyboard in Resources */ = {isa = PBXBuildFile; fileRef = 7D5B360820E28EEA0022BAE6 /* LaunchScreen.storyboard */; };
/* End PBXBuildFile section */

/* Begin PBXContainerItemProxy section */
		7D0342F620F4BA280050B6A6 /* PBXContainerItemProxy */ = {
			isa = PBXContainerItemProxy;
			containerPortal = 7D5B35F420E28EE80022BAE6 /* Project object */;
			proxyType = 1;
			remoteGlobalIDString = 7D5B35FB20E28EE80022BAE6;
			remoteInfo = XcodeProj;
		};
		7D03431820F4BB070050B6A6 /* PBXContainerItemProxy */ = {
			isa = PBXContainerItemProxy;
			containerPortal = 7D5B35F420E28EE80022BAE6 /* Project object */;
			proxyType = 1;
			remoteGlobalIDString = 7D03430C20F4BB070050B6A6;
			remoteInfo = TodayExtension;
		};
/* End PBXContainerItemProxy section */

/* Begin PBXCopyFilesBuildPhase section */
		7D03431E20F4BB070050B6A6 /* Embed App Extensions */ = {
			isa = PBXCopyFilesBuildPhase;
			buildActionMask = 2147483647;
			dstPath = "";
			dstSubfolderSpec = 13;
			files = (
				7D03431A20F4BB070050B6A6 /* TodayExtension.appex in Embed App Extensions */,
			);
			name = "Embed App Extensions";
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXCopyFilesBuildPhase section */

/* Begin PBXFileReference section */
		7D0342F120F4BA280050B6A6 /* XcodeProjUITests.xctest */ = {isa = PBXFileReference; explicitFileType = wrapper.cfbundle; includeInIndex = 0; path = XcodeProjUITests.xctest; sourceTree = BUILT_PRODUCTS_DIR; };
		7D0342F320F4BA280050B6A6 /* XcodeProjUITests.swift */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.swift; path = XcodeProjUITests.swift; sourceTree = "<group>"; };
		7D0342F520F4BA280050B6A6 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		7D03430D20F4BB070050B6A6 /* TodayExtension.appex */ = {isa = PBXFileReference; explicitFileType = "wrapper.app-extension"; includeInIndex = 0; path = TodayExtension.appex; sourceTree = BUILT_PRODUCTS_DIR; };
		7D03430F20F4BB070050B6A6 /* NotificationCenter.framework */ = {isa = PBXFileReference; lastKnownFileType = wrapper.framework; name = NotificationCenter.framework; path = System/Library/Frameworks/NotificationCenter.framework; sourceTree = SDKROOT; };
		7D03431220F4BB070050B6A6 /* TodayViewController.swift */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.swift; path = TodayViewController.swift; sourceTree = "<group>"; };
		7D03431520F4BB070050B6A6 /* Base */ = {isa = PBXFileReference; lastKnownFileType = file.storyboard; name = Base; path = Base.lproj/MainInterface.storyboard; sourceTree = "<group>"; };
		7D03431720F4BB070050B6A6 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		7D03431F20F4BB4A0050B6A6 /* TodayExtension.entitlements */ = {isa = PBXFileReference; lastKnownFileType = text.plist.entitlements; path = TodayExtension.entitlements; sourceTree = "<group>"; };
		7D03432020F4BB8D0050B6A6 /* CloudKit.framework */ = {isa = PBXFileReference; lastKnownFileType = wrapper.framework; name = CloudKit.framework; path = System/Library/Frameworks/CloudKit.framework; sourceTree = SDKROOT; };
		7D5B35FC20E28EE80022BAE6 /* XcodeProj.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = XcodeProj.app; sourceTree = BUILT_PRODUCTS_DIR; };
		7D5B35FF20E28EE80022BAE6 /* AppDelegate.swift */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.swift; path = AppDelegate.swift; sourceTree = "<group>"; };
		7D5B360120E28EE80022BAE6 /* ViewController.swift */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.swift; path = ViewController.swift; sourceTree = "<group>"; };
		7D5B360420E28EE80022BAE6 /* Base */ = {isa = PBXFileReference; lastKnownFileType = file.storyboard; name = Base; path = Base.lproj/Main.storyboard; sourceTree = "<group>"; };
		7D5B360620E28EEA0022BAE6 /* Assets.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Assets.xcassets; sourceTree = "<group>"; };
		7D5B360920E28EEA0022BAE6 /* Base */ = {isa = PBXFileReference; lastKnownFileType = file.storyboard; name = Base; path = Base.lproj/LaunchScreen.storyboard; sourceTree = "<group>"; };
		7D5B360B20E28EEA0022BAE6 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
/* End PBXFileReference section */

/* Begin PBXFrameworksBuildPhase section */
		7D0342EE20F4BA280050B6A6 /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		7D03430A20F4BB070050B6A6 /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
				7D03431020F4BB070050B6A6 /* NotificationCenter.framework in Frameworks */,
				7D03432120F4BB8D0050B6A6 /* CloudKit.framework in Frameworks */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		7D5B35F920E28EE80022BAE6 /* Frameworks */ = {
			isa = PBXFrameworksBuildPhase;
			buildActionMask = 2147483647;
			files = (
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXFrameworksBuildPhase section */

/* Begin PBXGroup section */
		7D0342F220F4BA280050B6A6 /* XcodeProjUITests */ = {
			isa = PBXGroup;
			children = (
				7D0342F320F4BA280050B6A6 /* XcodeProjUITests.swift */,
				7D0342F520F4BA280050B6A6 /* Info.plist */,
			);
			path = XcodeProjUITests;
			sourceTree = "<group>";
		};
		7D03430E20F4BB070050B6A6 /* Frameworks */ = {
			isa = PBXGroup;
			children = (
				7D03432020F4BB8D0050B6A6 /* CloudKit.framework */,
				7D03430F20F4BB070050B6A6 /* NotificationCenter.framework */,
			);
			name = Frameworks;
			sourceTree = "<group>";
		};
		7D03431120F4BB070050B6A6 /* TodayExtension */ = {
			isa = PBXGroup;
			children = (
				7D03431F20F4BB4A0050B6A6 /* TodayExtension.entitlements */,
				7D03431220F4BB070050B6A6 /* TodayViewController.swift */,
				7D03431420F4BB070050B6A6 /* MainInterface.storyboard */,
				7D03431720F4BB070050B6A6 /* Info.plist */,
			);
			path = TodayExtension;
			sourceTree = "<group>";
		};
		7D5B35F320E28EE80022BAE6 = {
			isa = PBXGroup;
			children = (
				7D5B35FE20E28EE80022BAE6 /* XcodeProj */,
				7D0342F220F4BA280050B6A6 /* XcodeProjUITests */,
				7D03431120F4BB070050B6A6 /* TodayExtension */,
				7D03430E20F4BB070050B6A6 /* Frameworks */,
				7D5B35FD20E28EE80022BAE6 /* Products */,
			);
			sourceTree = "<group>";
		};
		7D5B35FD20E28EE80022BAE6 /* Products */ = {
			isa = PBXGroup;
			children = (
				7D5B35FC20E28EE80022BAE6 /* XcodeProj.app */,
				7D0342F120F4BA280050B6A6 /* XcodeProjUITests.xctest */,
				7D03430D20F4BB070050B6A6 /* TodayExtension.appex */,
			);
			name = Products;
			sourceTree = "<group>";
		};
		7D5B35FE20E28EE80022BAE6 /* XcodeProj */ = {
			isa = PBXGroup;
			children = (
				7D5B35FF20E28EE80022BAE6 /* AppDelegate.swift */,
				7D5B360120E28EE80022BAE6 /* ViewController.swift */,
				7D5B360320E28EE80022BAE6 /* Main.storyboard */,
				7D5B360620E28EEA0022BAE6 /* Assets.xcassets */,
				7D5B360820E28EEA0022BAE6 /* LaunchScreen.storyboard */,
				7D5B360B20E28EEA0022BAE6 /* Info.plist */,
			);
			path = XcodeProj;
			sourceTree = "<group>";
		};
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
		7D0342F020F4BA280050B6A6 /* XcodeProjUITests */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 7D0342FA20F4BA280050B6A6 /* Build configuration list for PBXNativeTarget "XcodeProjUITests" */;
			buildPhases = (
				7D0342ED20F4BA280050B6A6 /* Sources */,
				7D0342EE20F4BA280050B6A6 /* Frameworks */,
				7D0342EF20F4BA280050B6A6 /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
				7D0342F720F4BA280050B6A6 /* PBXTargetDependency */,
			);
			name = XcodeProjUITests;
			productName = XcodeProjUITests;
			productReference = 7D0342F120F4BA280050B6A6 /* XcodeProjUITests.xctest */;
			productType = "com.apple.product-type.bundle.ui-testing";
		};
		7D03430C20F4BB070050B6A6 /* TodayExtension */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 7D03431B20F4BB070050B6A6 /* Build configuration list for PBXNativeTarget "TodayExtension" */;
			buildPhases = (
				7D03430920F4BB070050B6A6 /* Sources */,
				7D03430A20F4BB070050B6A6 /* Frameworks */,
				7D03430B20F4BB070050B6A6 /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = TodayExtension;
			productName = TodayExtension;
			productReference = 7D03430D20F4BB070050B6A6 /* TodayExtension.appex */;
			productType = "com.apple.product-type.app-extension";
		};
		7D5B35FB20E28EE80022BAE6 /* XcodeProj */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 7D5B360E20E28EEA0022BAE6 /* Build configuration list for PBXNativeTarget "XcodeProj" */;
			buildPhases = (
				7D5B35F820E28EE80022BAE6 /* Sources */,
				7D5B35F920E28EE80022BAE6 /* Frameworks */,
				7D5B35FA20E28EE80022BAE6 /* Resources */,
				7D03431E20F4BB070050B6A6 /* Embed App Extensions */,
			);
			buildRules = (
			);
			dependencies = (
				7D03431920F4BB070050B6A6 /* PBXTargetDependency */,
			);
			name = XcodeProj;
			productName = XcodeProj;
			productReference = 7D5B35FC20E28EE80022BAE6 /* XcodeProj.app */;
			productType = "com.apple.product-type.application";
		};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
		7D5B35F420E28EE80022BAE6 /* Project object */ = {
	attributes = {
		LastSwiftUpdateCheck = 0940;
		LastUpgradeCheck = 0940;
		ORGANIZATIONNAME = Bitrise;
		TargetAttributes = {
			7D0342F020F4BA280050B6A6 = {
				CreatedOnToolsVersion = "9.4.1";
				TestTargetID = 7D5B35FB20E28EE80022BAE6;
			};
			7D03430C20F4BB070050B6A6 = {
				CreatedOnToolsVersion = "9.4.1";
				SystemCapabilities = {
					"com.apple.Push" = {
						enabled = 1;
					};
					"com.apple.iCloud" = {
						enabled = 1;
					};
				};
			};
			7D5B35FB20E28EE80022BAE6 = {
				CreatedOnToolsVersion = "9.4.1";
				DevelopmentTeam = ABCD1234;
				DevelopmentTeamName = "";
				ProvisioningStyle = Manual;
			};
		};
	};
	buildConfigurationList = 7D5B35F720E28EE80022BAE6;
	compatibilityVersion = "Xcode 9.3";
	developmentRegion = en;
	hasScannedForEncodings = 0;
	isa = PBXProject;
	knownRegions = (
		en,
		Base,
	);
	mainGroup = 7D5B35F320E28EE80022BAE6;
	productRefGroup = 7D5B35FD20E28EE80022BAE6;
	projectDirPath = "";
	projectRoot = "";
	targets = (
		7D5B35FB20E28EE80022BAE6,
		7D0342F020F4BA280050B6A6,
		7D03430C20F4BB070050B6A6,
	);
};
/* End PBXProject section */

/* Begin PBXResourcesBuildPhase section */
		7D0342EF20F4BA280050B6A6 /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		7D03430B20F4BB070050B6A6 /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				7D03431620F4BB070050B6A6 /* MainInterface.storyboard in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		7D5B35FA20E28EE80022BAE6 /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				7D5B360A20E28EEA0022BAE6 /* LaunchScreen.storyboard in Resources */,
				7D5B360720E28EEA0022BAE6 /* Assets.xcassets in Resources */,
				7D5B360520E28EE80022BAE6 /* Main.storyboard in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXResourcesBuildPhase section */

/* Begin PBXSourcesBuildPhase section */
		7D0342ED20F4BA280050B6A6 /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				7D0342F420F4BA280050B6A6 /* XcodeProjUITests.swift in Sources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		7D03430920F4BB070050B6A6 /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				7D03431320F4BB070050B6A6 /* TodayViewController.swift in Sources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
		7D5B35F820E28EE80022BAE6 /* Sources */ = {
			isa = PBXSourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				7D5B360220E28EE80022BAE6 /* ViewController.swift in Sources */,
				7D5B360020E28EE80022BAE6 /* AppDelegate.swift in Sources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXSourcesBuildPhase section */

/* Begin PBXTargetDependency section */
		7D0342F720F4BA280050B6A6 /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 7D5B35FB20E28EE80022BAE6 /* XcodeProj */;
			targetProxy = 7D0342F620F4BA280050B6A6 /* PBXContainerItemProxy */;
		};
		7D03431920F4BB070050B6A6 /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 7D03430C20F4BB070050B6A6 /* TodayExtension */;
			targetProxy = 7D03431820F4BB070050B6A6 /* PBXContainerItemProxy */;
		};
/* End PBXTargetDependency section */

/* Begin PBXVariantGroup section */
		7D03431420F4BB070050B6A6 /* MainInterface.storyboard */ = {
			isa = PBXVariantGroup;
			children = (
				7D03431520F4BB070050B6A6 /* Base */,
			);
			name = MainInterface.storyboard;
			sourceTree = "<group>";
		};
		7D5B360320E28EE80022BAE6 /* Main.storyboard */ = {
			isa = PBXVariantGroup;
			children = (
				7D5B360420E28EE80022BAE6 /* Base */,
			);
			name = Main.storyboard;
			sourceTree = "<group>";
		};
		7D5B360820E28EEA0022BAE6 /* LaunchScreen.storyboard */ = {
			isa = PBXVariantGroup;
			children = (
				7D5B360920E28EEA0022BAE6 /* Base */,
			);
			name = LaunchScreen.storyboard;
			sourceTree = "<group>";
		};
/* End PBXVariantGroup section */

/* Begin XCBuildConfiguration section */
		7D0342F820F4BA280050B6A6 /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_STYLE = Automatic;
				DEVELOPMENT_TEAM = 72SA8V3WYL;
				INFOPLIST_FILE = XcodeProjUITests/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@loader_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = com.bitrise.XcodeProjUITests;
				PRODUCT_NAME = "$(TARGET_NAME)";
				SWIFT_VERSION = 4.0;
				TARGETED_DEVICE_FAMILY = "1,2";
				TEST_TARGET_NAME = XcodeProj;
			};
			name = Debug;
		};
		7D0342F920F4BA280050B6A6 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_STYLE = Automatic;
				DEVELOPMENT_TEAM = 72SA8V3WYL;
				INFOPLIST_FILE = XcodeProjUITests/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@loader_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = com.bitrise.XcodeProjUITests;
				PRODUCT_NAME = "$(TARGET_NAME)";
				SWIFT_VERSION = 4.0;
				TARGETED_DEVICE_FAMILY = "1,2";
				TEST_TARGET_NAME = XcodeProj;
			};
			name = Release;
		};
		7D03431C20F4BB070050B6A6 /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_ENTITLEMENTS = TodayExtension/TodayExtension.entitlements;
				CODE_SIGN_STYLE = Automatic;
				DEVELOPMENT_TEAM = 72SA8V3WYL;
				INFOPLIST_FILE = TodayExtension/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@executable_path/../../Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = com.bitrise.XcodeProj.TodayExtension;
				PRODUCT_NAME = "$(TARGET_NAME)";
				SKIP_INSTALL = YES;
				SWIFT_VERSION = 4.0;
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Debug;
		};
		7D03431D20F4BB070050B6A6 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				CODE_SIGN_ENTITLEMENTS = TodayExtension/TodayExtension.entitlements;
				CODE_SIGN_STYLE = Automatic;
				DEVELOPMENT_TEAM = 72SA8V3WYL;
				INFOPLIST_FILE = TodayExtension/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
					"@executable_path/../../Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = com.bitrise.XcodeProj.TodayExtension;
				PRODUCT_NAME = "$(TARGET_NAME)";
				SKIP_INSTALL = YES;
				SWIFT_VERSION = 4.0;
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Release;
		};
		7D5B360C20E28EEA0022BAE6 /* Debug */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_ANALYZER_NONNULL = YES;
				CLANG_ANALYZER_NUMBER_OBJECT_CONVERSION = YES_AGGRESSIVE;
				CLANG_CXX_LANGUAGE_STANDARD = "gnu++14";
				CLANG_CXX_LIBRARY = "libc++";
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CLANG_ENABLE_OBJC_WEAK = YES;
				CLANG_WARN_BLOCK_CAPTURE_AUTORELEASING = YES;
				CLANG_WARN_BOOL_CONVERSION = YES;
				CLANG_WARN_COMMA = YES;
				CLANG_WARN_CONSTANT_CONVERSION = YES;
				CLANG_WARN_DEPRECATED_OBJC_IMPLEMENTATIONS = YES;
				CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
				CLANG_WARN_DOCUMENTATION_COMMENTS = YES;
				CLANG_WARN_EMPTY_BODY = YES;
				CLANG_WARN_ENUM_CONVERSION = YES;
				CLANG_WARN_INFINITE_RECURSION = YES;
				CLANG_WARN_INT_CONVERSION = YES;
				CLANG_WARN_NON_LITERAL_NULL_CONVERSION = YES;
				CLANG_WARN_OBJC_IMPLICIT_RETAIN_SELF = YES;
				CLANG_WARN_OBJC_LITERAL_CONVERSION = YES;
				CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
				CLANG_WARN_RANGE_LOOP_ANALYSIS = YES;
				CLANG_WARN_STRICT_PROTOTYPES = YES;
				CLANG_WARN_SUSPICIOUS_MOVE = YES;
				CLANG_WARN_UNGUARDED_AVAILABILITY = YES_AGGRESSIVE;
				CLANG_WARN_UNREACHABLE_CODE = YES;
				CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
				CODE_SIGN_IDENTITY = "iPhone Developer";
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = dwarf;
				ENABLE_STRICT_OBJC_MSGSEND = YES;
				ENABLE_TESTABILITY = YES;
				GCC_C_LANGUAGE_STANDARD = gnu11;
				GCC_DYNAMIC_NO_PIC = NO;
				GCC_NO_COMMON_BLOCKS = YES;
				GCC_OPTIMIZATION_LEVEL = 0;
				GCC_PREPROCESSOR_DEFINITIONS = (
					"DEBUG=1",
					"$(inherited)",
				);
				GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
				GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
				GCC_WARN_UNDECLARED_SELECTOR = YES;
				GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
				GCC_WARN_UNUSED_FUNCTION = YES;
				GCC_WARN_UNUSED_VARIABLE = YES;
				IPHONEOS_DEPLOYMENT_TARGET = 11.4;
				MTL_ENABLE_DEBUG_INFO = YES;
				ONLY_ACTIVE_ARCH = YES;
				SDKROOT = iphoneos;
				SWIFT_ACTIVE_COMPILATION_CONDITIONS = DEBUG;
				SWIFT_OPTIMIZATION_LEVEL = "-Onone";
			};
			name = Debug;
		};
		7D5B360D20E28EEA0022BAE6 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_ANALYZER_NONNULL = YES;
				CLANG_ANALYZER_NUMBER_OBJECT_CONVERSION = YES_AGGRESSIVE;
				CLANG_CXX_LANGUAGE_STANDARD = "gnu++14";
				CLANG_CXX_LIBRARY = "libc++";
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CLANG_ENABLE_OBJC_WEAK = YES;
				CLANG_WARN_BLOCK_CAPTURE_AUTORELEASING = YES;
				CLANG_WARN_BOOL_CONVERSION = YES;
				CLANG_WARN_COMMA = YES;
				CLANG_WARN_CONSTANT_CONVERSION = YES;
				CLANG_WARN_DEPRECATED_OBJC_IMPLEMENTATIONS = YES;
				CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
				CLANG_WARN_DOCUMENTATION_COMMENTS = YES;
				CLANG_WARN_EMPTY_BODY = YES;
				CLANG_WARN_ENUM_CONVERSION = YES;
				CLANG_WARN_INFINITE_RECURSION = YES;
				CLANG_WARN_INT_CONVERSION = YES;
				CLANG_WARN_NON_LITERAL_NULL_CONVERSION = YES;
				CLANG_WARN_OBJC_IMPLICIT_RETAIN_SELF = YES;
				CLANG_WARN_OBJC_LITERAL_CONVERSION = YES;
				CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
				CLANG_WARN_RANGE_LOOP_ANALYSIS = YES;
				CLANG_WARN_STRICT_PROTOTYPES = YES;
				CLANG_WARN_SUSPICIOUS_MOVE = YES;
				CLANG_WARN_UNGUARDED_AVAILABILITY = YES_AGGRESSIVE;
				CLANG_WARN_UNREACHABLE_CODE = YES;
				CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
				CODE_SIGN_IDENTITY = "iPhone Developer";
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
				ENABLE_NS_ASSERTIONS = NO;
				ENABLE_STRICT_OBJC_MSGSEND = YES;
				GCC_C_LANGUAGE_STANDARD = gnu11;
				GCC_NO_COMMON_BLOCKS = YES;
				GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
				GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
				GCC_WARN_UNDECLARED_SELECTOR = YES;
				GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
				GCC_WARN_UNUSED_FUNCTION = YES;
				GCC_WARN_UNUSED_VARIABLE = YES;
				IPHONEOS_DEPLOYMENT_TARGET = 11.4;
				MTL_ENABLE_DEBUG_INFO = NO;
				SDKROOT = iphoneos;
				SWIFT_COMPILATION_MODE = wholemodule;
				SWIFT_OPTIMIZATION_LEVEL = "-O";
				VALIDATE_PRODUCT = YES;
			};
			name = Release;
		};
		7D5B360F20E28EEA0022BAE6 /* Debug */ = {
	buildSettings = {
		"ALWAYS_EMBED_SWIFT_STANDARD_LIBRARIES" = YES;
		"ASSETCATALOG_COMPILER_APPICON_NAME" = AppIcon;
		"CODE_SIGN_IDENTITY" = "Apple Development: John Doe (ASDF1234)";
		"CODE_SIGN_STYLE" = Manual;
		"DEVELOPMENT_TEAM" = ABCD1234;
		"INFOPLIST_FILE" = "XcodeProj/Info.plist";
		"LD_RUNPATH_SEARCH_PATHS" = (
			"$(inherited)",
			"@executable_path/Frameworks",
		);
		"PRODUCT_BUNDLE_IDENTIFIER" = "com.bitrise.XcodeProj";
		"PRODUCT_NAME" = "$(TARGET_NAME)";
		"PROVISIONING_PROFILE" = "asdf56b6-e75a-4f86-bf25-101bfc2fasdf";
		"PROVISIONING_PROFILE_SPECIFIER" = "";
		"SWIFT_VERSION" = "4.0";
		"TARGETED_DEVICE_FAMILY" = "1,2";
	};
	isa = XCBuildConfiguration;
	name = Debug;
};
		7D5B361020E28EEA0022BAE6 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_EMBED_SWIFT_STANDARD_LIBRARIES = YES;
				ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
				CODE_SIGN_STYLE = Automatic;
				DEVELOPMENT_TEAM = 72SA8V3WYL;
				INFOPLIST_FILE = XcodeProj/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = (
					"$(inherited)",
					"@executable_path/Frameworks",
				);
				PRODUCT_BUNDLE_IDENTIFIER = com.bitrise.XcodeProj;
				PRODUCT_NAME = "$(TARGET_NAME)";
				SWIFT_VERSION = 4.0;
				TARGETED_DEVICE_FAMILY = "1,2";
			};
			name = Release;
		};
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
		7D0342FA20F4BA280050B6A6 /* Build configuration list for PBXNativeTarget "XcodeProjUITests" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				7D0342F820F4BA280050B6A6 /* Debug */,
				7D0342F920F4BA280050B6A6 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		7D03431B20F4BB070050B6A6 /* Build configuration list for PBXNativeTarget "TodayExtension" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				7D03431C20F4BB070050B6A6 /* Debug */,
				7D03431D20F4BB070050B6A6 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		7D5B35F720E28EE80022BAE6 /* Build configuration list for PBXProject "XcodeProj" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				7D5B360C20E28EEA0022BAE6 /* Debug */,
				7D5B360D20E28EEA0022BAE6 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		7D5B360E20E28EEA0022BAE6 /* Build configuration list for PBXNativeTarget "XcodeProj" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				7D5B360F20E28EEA0022BAE6 /* Debug */,
				7D5B361020E28EEA0022BAE6 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
/* End XCConfigurationList section */
	};
	rootObject = 7D5B35F420E28EE80022BAE6 /* Project object */;
}
`
