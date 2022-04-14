package xcodeproj

import (
	"fmt"
	"testing"

	plist "github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/pretty"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/stretchr/testify/require"
)

func TestParseProjWithoutUnitTestProductType(t *testing.T) {
	var raw serialized.Object
	_, err := plist.Unmarshal([]byte(rawProj), &raw)
	require.NoError(t, err)

	{
		proj, err := parseProj("BA3CBE6D19F7A93800CED4D5", raw, nil)
		require.NoError(t, err)
		fmt.Printf("proj:\n%s\n", pretty.Object(proj))
		require.Equal(t, expectedProj, pretty.Object(proj))
	}

	{
		proj, err := parseProj("INVALID_TARGET_ID", raw, nil)
		require.Error(t, err)
		require.Equal(t, Proj{}, proj)
	}
}

const rawProj = `
{
	/* Begin PBXBuildFile section */
			BA3CBE7B19F7A93800CED4D5 /* main.m in Sources */ = {isa = PBXBuildFile; fileRef = BA3CBE7A19F7A93800CED4D5 /* main.m */; };
			BA3CBE7E19F7A93900CED4D5 /* AppDelegate.m in Sources */ = {isa = PBXBuildFile; fileRef = BA3CBE7D19F7A93900CED4D5 /* AppDelegate.m */; };
			BA3CBE8119F7A93900CED4D5 /* ios_simple_objc.xcdatamodeld in Sources */ = {isa = PBXBuildFile; fileRef = BA3CBE7F19F7A93900CED4D5 /* ios_simple_objc.xcdatamodeld */; };
			BA3CBE8419F7A93900CED4D5 /* ViewController.m in Sources */ = {isa = PBXBuildFile; fileRef = BA3CBE8319F7A93900CED4D5 /* ViewController.m */; };
			BA3CBE8719F7A93900CED4D5 /* Main.storyboard in Resources */ = {isa = PBXBuildFile; fileRef = BA3CBE8519F7A93900CED4D5 /* Main.storyboard */; };
			BA3CBE8919F7A93900CED4D5 /* Images.xcassets in Resources */ = {isa = PBXBuildFile; fileRef = BA3CBE8819F7A93900CED4D5 /* Images.xcassets */; };
			BA3CBE8C19F7A93900CED4D5 /* LaunchScreen.xib in Resources */ = {isa = PBXBuildFile; fileRef = BA3CBE8A19F7A93900CED4D5 /* LaunchScreen.xib */; };
			BA3CBE9819F7A93900CED4D5 /* ios_simple_objcTests.m in Sources */ = {isa = PBXBuildFile; fileRef = BA3CBE9719F7A93900CED4D5 /* ios_simple_objcTests.m */; };
	/* End PBXBuildFile section */
	
	/* Begin PBXContainerItemProxy section */
			BA3CBE9219F7A93900CED4D5 /* PBXContainerItemProxy */ = {
				isa = PBXContainerItemProxy;
				containerPortal = BA3CBE6D19F7A93800CED4D5 /* Project object */;
				proxyType = 1;
				remoteGlobalIDString = BA3CBE7419F7A93800CED4D5;
				remoteInfo = "ios-simple-objc";
			};
	/* End PBXContainerItemProxy section */
	
	/* Begin PBXFileReference section */
			BA3CBE7519F7A93800CED4D5 /* ios-simple-objc.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = "ios-simple-objc.app"; sourceTree = BUILT_PRODUCTS_DIR; };
			BA3CBE7919F7A93800CED4D5 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
			BA3CBE7A19F7A93800CED4D5 /* main.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = main.m; sourceTree = "<group>"; };
			BA3CBE7C19F7A93800CED4D5 /* AppDelegate.h */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.h; path = AppDelegate.h; sourceTree = "<group>"; };
			BA3CBE7D19F7A93900CED4D5 /* AppDelegate.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = AppDelegate.m; sourceTree = "<group>"; };
			BA3CBE8019F7A93900CED4D5 /* ios_simple_objc.xcdatamodel */ = {isa = PBXFileReference; lastKnownFileType = wrapper.xcdatamodel; path = ios_simple_objc.xcdatamodel; sourceTree = "<group>"; };
			BA3CBE8219F7A93900CED4D5 /* ViewController.h */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.h; path = ViewController.h; sourceTree = "<group>"; };
			BA3CBE8319F7A93900CED4D5 /* ViewController.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = ViewController.m; sourceTree = "<group>"; };
			BA3CBE8619F7A93900CED4D5 /* Base */ = {isa = PBXFileReference; lastKnownFileType = file.storyboard; name = Base; path = Base.lproj/Main.storyboard; sourceTree = "<group>"; };
			BA3CBE8819F7A93900CED4D5 /* Images.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Images.xcassets; sourceTree = "<group>"; };
			BA3CBE8B19F7A93900CED4D5 /* Base */ = {isa = PBXFileReference; lastKnownFileType = file.xib; name = Base; path = Base.lproj/LaunchScreen.xib; sourceTree = "<group>"; };
			BA3CBE9119F7A93900CED4D5 /* ios-simple-objcTests.xctest */ = {isa = PBXFileReference; explicitFileType = wrapper.cfbundle; includeInIndex = 0; path = "ios-simple-objcTests.xctest"; sourceTree = BUILT_PRODUCTS_DIR; };
			BA3CBE9619F7A93900CED4D5 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
			BA3CBE9719F7A93900CED4D5 /* ios_simple_objcTests.m */ = {isa = PBXFileReference; lastKnownFileType = sourcecode.c.objc; path = ios_simple_objcTests.m; sourceTree = "<group>"; };
	/* End PBXFileReference section */
	
	/* Begin PBXFrameworksBuildPhase section */
			BA3CBE7219F7A93800CED4D5 /* Frameworks */ = {
				isa = PBXFrameworksBuildPhase;
				buildActionMask = 2147483647;
				files = (
				);
				runOnlyForDeploymentPostprocessing = 0;
			};
			BA3CBE8E19F7A93900CED4D5 /* Frameworks */ = {
				isa = PBXFrameworksBuildPhase;
				buildActionMask = 2147483647;
				files = (
				);
				runOnlyForDeploymentPostprocessing = 0;
			};
	/* End PBXFrameworksBuildPhase section */
	
	/* Begin PBXGroup section */
			BA3CBE6C19F7A93800CED4D5 = {
				isa = PBXGroup;
				children = (
					BA3CBE7719F7A93800CED4D5 /* ios-simple-objc */,
					BA3CBE9419F7A93900CED4D5 /* ios-simple-objcTests */,
					BA3CBE7619F7A93800CED4D5 /* Products */,
				);
				sourceTree = "<group>";
			};
			BA3CBE7619F7A93800CED4D5 /* Products */ = {
				isa = PBXGroup;
				children = (
					BA3CBE7519F7A93800CED4D5 /* ios-simple-objc.app */,
					BA3CBE9119F7A93900CED4D5 /* ios-simple-objcTests.xctest */,
				);
				name = Products;
				sourceTree = "<group>";
			};
			BA3CBE7719F7A93800CED4D5 /* ios-simple-objc */ = {
				isa = PBXGroup;
				children = (
					BA3CBE7C19F7A93800CED4D5 /* AppDelegate.h */,
					BA3CBE7D19F7A93900CED4D5 /* AppDelegate.m */,
					BA3CBE8219F7A93900CED4D5 /* ViewController.h */,
					BA3CBE8319F7A93900CED4D5 /* ViewController.m */,
					BA3CBE8519F7A93900CED4D5 /* Main.storyboard */,
					BA3CBE8819F7A93900CED4D5 /* Images.xcassets */,
					BA3CBE8A19F7A93900CED4D5 /* LaunchScreen.xib */,
					BA3CBE7F19F7A93900CED4D5 /* ios_simple_objc.xcdatamodeld */,
					BA3CBE7819F7A93800CED4D5 /* Supporting Files */,
				);
				path = "ios-simple-objc";
				sourceTree = "<group>";
			};
			BA3CBE7819F7A93800CED4D5 /* Supporting Files */ = {
				isa = PBXGroup;
				children = (
					BA3CBE7919F7A93800CED4D5 /* Info.plist */,
					BA3CBE7A19F7A93800CED4D5 /* main.m */,
				);
				name = "Supporting Files";
				sourceTree = "<group>";
			};
			BA3CBE9419F7A93900CED4D5 /* ios-simple-objcTests */ = {
				isa = PBXGroup;
				children = (
					BA3CBE9719F7A93900CED4D5 /* ios_simple_objcTests.m */,
					BA3CBE9519F7A93900CED4D5 /* Supporting Files */,
				);
				path = "ios-simple-objcTests";
				sourceTree = "<group>";
			};
			BA3CBE9519F7A93900CED4D5 /* Supporting Files */ = {
				isa = PBXGroup;
				children = (
					BA3CBE9619F7A93900CED4D5 /* Info.plist */,
				);
				name = "Supporting Files";
				sourceTree = "<group>";
			};
	/* End PBXGroup section */
	
	/* Begin PBXNativeTarget section */
			BA3CBE7419F7A93800CED4D5 /* ios-simple-objc */ = {
				isa = PBXNativeTarget;
				buildConfigurationList = BA3CBE9B19F7A93900CED4D5 /* Build configuration list for PBXNativeTarget "ios-simple-objc" */;
				buildPhases = (
					BA3CBE7119F7A93800CED4D5 /* Sources */,
					BA3CBE7219F7A93800CED4D5 /* Frameworks */,
					BA3CBE7319F7A93800CED4D5 /* Resources */,
				);
				buildRules = (
				);
				dependencies = (
				);
				name = "ios-simple-objc";
				productName = "ios-simple-objc";
				productReference = BA3CBE7519F7A93800CED4D5 /* ios-simple-objc.app */;
				productType = "com.apple.product-type.application";
			};
			BA3CBE9019F7A93900CED4D5 /* ios-simple-objcTests */ = {
				isa = PBXNativeTarget;
				buildConfigurationList = BA3CBE9E19F7A93900CED4D5 /* Build configuration list for PBXNativeTarget "ios-simple-objcTests" */;
				buildPhases = (
					BA3CBE8D19F7A93900CED4D5 /* Sources */,
					BA3CBE8E19F7A93900CED4D5 /* Frameworks */,
					BA3CBE8F19F7A93900CED4D5 /* Resources */,
				);
				buildRules = (
				);
				dependencies = (
					BA3CBE9319F7A93900CED4D5 /* PBXTargetDependency */,
				);
				name = "ios-simple-objcTests";
				productName = "ios-simple-objcTests";
				productReference = BA3CBE9119F7A93900CED4D5 /* ios-simple-objcTests.xctest */;
				productType = "com.apple.product-type.bundle.unit-test";
			};
	/* End PBXNativeTarget section */
	
	/* Begin PBXProject section */
			BA3CBE6D19F7A93800CED4D5 /* Project object */ = {
				isa = PBXProject;
				attributes = {
					LastUpgradeCheck = 0800;
					ORGANIZATIONNAME = Bitrise;
					TargetAttributes = {
						BA3CBE7419F7A93800CED4D5 = {
							CreatedOnToolsVersion = 6.1;
							DevelopmentTeam = 72SA8V3WYL;
							ProvisioningStyle = Manual;
						};
						BA3CBE9019F7A93900CED4D5 = {
							CreatedOnToolsVersion = 6.1;
							TestTargetID = BA3CBE7419F7A93800CED4D5;
						};
					};
				};
				buildConfigurationList = BA3CBE7019F7A93800CED4D5 /* Build configuration list for PBXProject "ios-simple-objc" */;
				compatibilityVersion = "Xcode 3.2";
				developmentRegion = English;
				hasScannedForEncodings = 0;
				knownRegions = (
					en,
					Base,
				);
				mainGroup = BA3CBE6C19F7A93800CED4D5;
				productRefGroup = BA3CBE7619F7A93800CED4D5 /* Products */;
				projectDirPath = "";
				projectRoot = "";
				targets = (
					BA3CBE7419F7A93800CED4D5 /* ios-simple-objc */,
					BA3CBE9019F7A93900CED4D5 /* ios-simple-objcTests */,
				);
			};
	/* End PBXProject section */
	
	/* Begin PBXResourcesBuildPhase section */
			BA3CBE7319F7A93800CED4D5 /* Resources */ = {
				isa = PBXResourcesBuildPhase;
				buildActionMask = 2147483647;
				files = (
					BA3CBE8719F7A93900CED4D5 /* Main.storyboard in Resources */,
					BA3CBE8C19F7A93900CED4D5 /* LaunchScreen.xib in Resources */,
					BA3CBE8919F7A93900CED4D5 /* Images.xcassets in Resources */,
				);
				runOnlyForDeploymentPostprocessing = 0;
			};
			BA3CBE8F19F7A93900CED4D5 /* Resources */ = {
				isa = PBXResourcesBuildPhase;
				buildActionMask = 2147483647;
				files = (
				);
				runOnlyForDeploymentPostprocessing = 0;
			};
	/* End PBXResourcesBuildPhase section */
	
	/* Begin PBXSourcesBuildPhase section */
			BA3CBE7119F7A93800CED4D5 /* Sources */ = {
				isa = PBXSourcesBuildPhase;
				buildActionMask = 2147483647;
				files = (
					BA3CBE7E19F7A93900CED4D5 /* AppDelegate.m in Sources */,
					BA3CBE7B19F7A93800CED4D5 /* main.m in Sources */,
					BA3CBE8419F7A93900CED4D5 /* ViewController.m in Sources */,
					BA3CBE8119F7A93900CED4D5 /* ios_simple_objc.xcdatamodeld in Sources */,
				);
				runOnlyForDeploymentPostprocessing = 0;
			};
			BA3CBE8D19F7A93900CED4D5 /* Sources */ = {
				isa = PBXSourcesBuildPhase;
				buildActionMask = 2147483647;
				files = (
					BA3CBE9819F7A93900CED4D5 /* ios_simple_objcTests.m in Sources */,
				);
				runOnlyForDeploymentPostprocessing = 0;
			};
	/* End PBXSourcesBuildPhase section */
	
	/* Begin PBXTargetDependency section */
			BA3CBE9319F7A93900CED4D5 /* PBXTargetDependency */ = {
				isa = PBXTargetDependency;
				target = BA3CBE7419F7A93800CED4D5 /* ios-simple-objc */;
				targetProxy = BA3CBE9219F7A93900CED4D5 /* PBXContainerItemProxy */;
			};
	/* End PBXTargetDependency section */
	
	/* Begin PBXVariantGroup section */
			BA3CBE8519F7A93900CED4D5 /* Main.storyboard */ = {
				isa = PBXVariantGroup;
				children = (
					BA3CBE8619F7A93900CED4D5 /* Base */,
				);
				name = Main.storyboard;
				sourceTree = "<group>";
			};
			BA3CBE8A19F7A93900CED4D5 /* LaunchScreen.xib */ = {
				isa = PBXVariantGroup;
				children = (
					BA3CBE8B19F7A93900CED4D5 /* Base */,
				);
				name = LaunchScreen.xib;
				sourceTree = "<group>";
			};
	/* End PBXVariantGroup section */
	
	/* Begin XCBuildConfiguration section */
			BA3CBE9919F7A93900CED4D5 /* Debug */ = {
				isa = XCBuildConfiguration;
				buildSettings = {
					ALWAYS_SEARCH_USER_PATHS = NO;
					CLANG_CXX_LANGUAGE_STANDARD = "gnu++0x";
					CLANG_CXX_LIBRARY = "libc++";
					CLANG_ENABLE_MODULES = YES;
					CLANG_ENABLE_OBJC_ARC = YES;
					CLANG_WARN_BOOL_CONVERSION = YES;
					CLANG_WARN_CONSTANT_CONVERSION = YES;
					CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
					CLANG_WARN_EMPTY_BODY = YES;
					CLANG_WARN_ENUM_CONVERSION = YES;
					CLANG_WARN_INFINITE_RECURSION = YES;
					CLANG_WARN_INT_CONVERSION = YES;
					CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
					CLANG_WARN_SUSPICIOUS_MOVE = YES;
					CLANG_WARN_UNREACHABLE_CODE = YES;
					CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
					"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
					COPY_PHASE_STRIP = NO;
					ENABLE_STRICT_OBJC_MSGSEND = YES;
					ENABLE_TESTABILITY = YES;
					GCC_C_LANGUAGE_STANDARD = gnu99;
					GCC_DYNAMIC_NO_PIC = NO;
					GCC_NO_COMMON_BLOCKS = YES;
					GCC_OPTIMIZATION_LEVEL = 0;
					GCC_PREPROCESSOR_DEFINITIONS = (
						"DEBUG=1",
						"$(inherited)",
					);
					GCC_SYMBOLS_PRIVATE_EXTERN = NO;
					GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
					GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
					GCC_WARN_UNDECLARED_SELECTOR = YES;
					GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
					GCC_WARN_UNUSED_FUNCTION = YES;
					GCC_WARN_UNUSED_VARIABLE = YES;
					IPHONEOS_DEPLOYMENT_TARGET = 8.1;
					MTL_ENABLE_DEBUG_INFO = YES;
					ONLY_ACTIVE_ARCH = YES;
					SDKROOT = iphoneos;
					TARGETED_DEVICE_FAMILY = "1,2";
				};
				name = Debug;
			};
			BA3CBE9A19F7A93900CED4D5 /* Release */ = {
				isa = XCBuildConfiguration;
				buildSettings = {
					ALWAYS_SEARCH_USER_PATHS = NO;
					CLANG_CXX_LANGUAGE_STANDARD = "gnu++0x";
					CLANG_CXX_LIBRARY = "libc++";
					CLANG_ENABLE_MODULES = YES;
					CLANG_ENABLE_OBJC_ARC = YES;
					CLANG_WARN_BOOL_CONVERSION = YES;
					CLANG_WARN_CONSTANT_CONVERSION = YES;
					CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
					CLANG_WARN_EMPTY_BODY = YES;
					CLANG_WARN_ENUM_CONVERSION = YES;
					CLANG_WARN_INFINITE_RECURSION = YES;
					CLANG_WARN_INT_CONVERSION = YES;
					CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
					CLANG_WARN_SUSPICIOUS_MOVE = YES;
					CLANG_WARN_UNREACHABLE_CODE = YES;
					CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
					"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
					COPY_PHASE_STRIP = YES;
					ENABLE_NS_ASSERTIONS = NO;
					ENABLE_STRICT_OBJC_MSGSEND = YES;
					GCC_C_LANGUAGE_STANDARD = gnu99;
					GCC_NO_COMMON_BLOCKS = YES;
					GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
					GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
					GCC_WARN_UNDECLARED_SELECTOR = YES;
					GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
					GCC_WARN_UNUSED_FUNCTION = YES;
					GCC_WARN_UNUSED_VARIABLE = YES;
					IPHONEOS_DEPLOYMENT_TARGET = 8.1;
					MTL_ENABLE_DEBUG_INFO = NO;
					SDKROOT = iphoneos;
					TARGETED_DEVICE_FAMILY = "1,2";
					VALIDATE_PRODUCT = YES;
				};
				name = Release;
			};
			BA3CBE9C19F7A93900CED4D5 /* Debug */ = {
				isa = XCBuildConfiguration;
				buildSettings = {
					ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
					CODE_SIGN_IDENTITY = "iPhone Developer";
					"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
					DEVELOPMENT_TEAM = 72SA8V3WYL;
					INFOPLIST_FILE = "ios-simple-objc/Info.plist";
					LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
					PRODUCT_BUNDLE_IDENTIFIER = "Bitrise.$(PRODUCT_NAME:rfc1034identifier)";
					PRODUCT_NAME = "$(TARGET_NAME)";
					PROVISIONING_PROFILE = "";
					PROVISIONING_PROFILE_SPECIFIER = "BitriseBot-Wildcard";
				};
				name = Debug;
			};
			BA3CBE9D19F7A93900CED4D5 /* Release */ = {
				isa = XCBuildConfiguration;
				buildSettings = {
					ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
					CODE_SIGN_IDENTITY = "iPhone Developer";
					"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
					DEVELOPMENT_TEAM = 72SA8V3WYL;
					INFOPLIST_FILE = "ios-simple-objc/Info.plist";
					LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
					PRODUCT_BUNDLE_IDENTIFIER = "Bitrise.$(PRODUCT_NAME:rfc1034identifier)";
					PRODUCT_NAME = "$(TARGET_NAME)";
					PROVISIONING_PROFILE = "";
					PROVISIONING_PROFILE_SPECIFIER = "BitriseBot-Wildcard";
				};
				name = Release;
			};
			BA3CBE9F19F7A93900CED4D5 /* Debug */ = {
				isa = XCBuildConfiguration;
				buildSettings = {
					BUNDLE_LOADER = "$(TEST_HOST)";
					FRAMEWORK_SEARCH_PATHS = (
						"$(SDKROOT)/Developer/Library/Frameworks",
						"$(inherited)",
					);
					GCC_PREPROCESSOR_DEFINITIONS = (
						"DEBUG=1",
						"$(inherited)",
					);
					INFOPLIST_FILE = "ios-simple-objcTests/Info.plist";
					LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks @loader_path/Frameworks";
					PRODUCT_BUNDLE_IDENTIFIER = "Bitrise.$(PRODUCT_NAME:rfc1034identifier)";
					PRODUCT_NAME = "$(TARGET_NAME)";
					TEST_HOST = "$(BUILT_PRODUCTS_DIR)/ios-simple-objc.app/ios-simple-objc";
				};
				name = Debug;
			};
			BA3CBEA019F7A93900CED4D5 /* Release */ = {
				isa = XCBuildConfiguration;
				buildSettings = {
					BUNDLE_LOADER = "$(TEST_HOST)";
					FRAMEWORK_SEARCH_PATHS = (
						"$(SDKROOT)/Developer/Library/Frameworks",
						"$(inherited)",
					);
					INFOPLIST_FILE = "ios-simple-objcTests/Info.plist";
					LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks @loader_path/Frameworks";
					PRODUCT_BUNDLE_IDENTIFIER = "Bitrise.$(PRODUCT_NAME:rfc1034identifier)";
					PRODUCT_NAME = "$(TARGET_NAME)";
					TEST_HOST = "$(BUILT_PRODUCTS_DIR)/ios-simple-objc.app/ios-simple-objc";
				};
				name = Release;
			};
	/* End XCBuildConfiguration section */
	
	/* Begin XCConfigurationList section */
			BA3CBE7019F7A93800CED4D5 /* Build configuration list for PBXProject "ios-simple-objc" */ = {
				isa = XCConfigurationList;
				buildConfigurations = (
					BA3CBE9919F7A93900CED4D5 /* Debug */,
					BA3CBE9A19F7A93900CED4D5 /* Release */,
				);
				defaultConfigurationIsVisible = 0;
				defaultConfigurationName = Release;
			};
			BA3CBE9B19F7A93900CED4D5 /* Build configuration list for PBXNativeTarget "ios-simple-objc" */ = {
				isa = XCConfigurationList;
				buildConfigurations = (
					BA3CBE9C19F7A93900CED4D5 /* Debug */,
					BA3CBE9D19F7A93900CED4D5 /* Release */,
				);
				defaultConfigurationIsVisible = 0;
				defaultConfigurationName = Release;
			};
			BA3CBE9E19F7A93900CED4D5 /* Build configuration list for PBXNativeTarget "ios-simple-objcTests" */ = {
				isa = XCConfigurationList;
				buildConfigurations = (
					BA3CBE9F19F7A93900CED4D5 /* Debug */,
					BA3CBEA019F7A93900CED4D5 /* Release */,
				);
				defaultConfigurationIsVisible = 0;
				defaultConfigurationName = Release;
			};
	/* End XCConfigurationList section */
	
	/* Begin XCVersionGroup section */
			BA3CBE7F19F7A93900CED4D5 /* ios_simple_objc.xcdatamodeld */ = {
				isa = XCVersionGroup;
				children = (
					BA3CBE8019F7A93900CED4D5 /* ios_simple_objc.xcdatamodel */,
				);
				currentVersion = BA3CBE8019F7A93900CED4D5 /* ios_simple_objc.xcdatamodel */;
				path = ios_simple_objc.xcdatamodeld;
				sourceTree = "<group>";
				versionGroupType = wrapper.xcdatamodel;
			};
	/* End XCVersionGroup section */
}
`

const expectedProj = `{
	"ID": "BA3CBE6D19F7A93800CED4D5",
	"BuildConfigurationList": {
		"ID": "BA3CBE7019F7A93800CED4D5",
		"DefaultConfigurationName": "Release",
		"BuildConfigurations": [
			{
				"ID": "BA3CBE9919F7A93900CED4D5",
				"Name": "Debug",
				"BuildSettings": {
					"ALWAYS_SEARCH_USER_PATHS": "NO",
					"CLANG_CXX_LANGUAGE_STANDARD": "gnu++0x",
					"CLANG_CXX_LIBRARY": "libc++",
					"CLANG_ENABLE_MODULES": "YES",
					"CLANG_ENABLE_OBJC_ARC": "YES",
					"CLANG_WARN_BOOL_CONVERSION": "YES",
					"CLANG_WARN_CONSTANT_CONVERSION": "YES",
					"CLANG_WARN_DIRECT_OBJC_ISA_USAGE": "YES_ERROR",
					"CLANG_WARN_EMPTY_BODY": "YES",
					"CLANG_WARN_ENUM_CONVERSION": "YES",
					"CLANG_WARN_INFINITE_RECURSION": "YES",
					"CLANG_WARN_INT_CONVERSION": "YES",
					"CLANG_WARN_OBJC_ROOT_CLASS": "YES_ERROR",
					"CLANG_WARN_SUSPICIOUS_MOVE": "YES",
					"CLANG_WARN_UNREACHABLE_CODE": "YES",
					"CLANG_WARN__DUPLICATE_METHOD_MATCH": "YES",
					"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
					"COPY_PHASE_STRIP": "NO",
					"ENABLE_STRICT_OBJC_MSGSEND": "YES",
					"ENABLE_TESTABILITY": "YES",
					"GCC_C_LANGUAGE_STANDARD": "gnu99",
					"GCC_DYNAMIC_NO_PIC": "NO",
					"GCC_NO_COMMON_BLOCKS": "YES",
					"GCC_OPTIMIZATION_LEVEL": "0",
					"GCC_PREPROCESSOR_DEFINITIONS": [
						"DEBUG=1",
						"$(inherited)"
					],
					"GCC_SYMBOLS_PRIVATE_EXTERN": "NO",
					"GCC_WARN_64_TO_32_BIT_CONVERSION": "YES",
					"GCC_WARN_ABOUT_RETURN_TYPE": "YES_ERROR",
					"GCC_WARN_UNDECLARED_SELECTOR": "YES",
					"GCC_WARN_UNINITIALIZED_AUTOS": "YES_AGGRESSIVE",
					"GCC_WARN_UNUSED_FUNCTION": "YES",
					"GCC_WARN_UNUSED_VARIABLE": "YES",
					"IPHONEOS_DEPLOYMENT_TARGET": "8.1",
					"MTL_ENABLE_DEBUG_INFO": "YES",
					"ONLY_ACTIVE_ARCH": "YES",
					"SDKROOT": "iphoneos",
					"TARGETED_DEVICE_FAMILY": "1,2"
				}
			},
			{
				"ID": "BA3CBE9A19F7A93900CED4D5",
				"Name": "Release",
				"BuildSettings": {
					"ALWAYS_SEARCH_USER_PATHS": "NO",
					"CLANG_CXX_LANGUAGE_STANDARD": "gnu++0x",
					"CLANG_CXX_LIBRARY": "libc++",
					"CLANG_ENABLE_MODULES": "YES",
					"CLANG_ENABLE_OBJC_ARC": "YES",
					"CLANG_WARN_BOOL_CONVERSION": "YES",
					"CLANG_WARN_CONSTANT_CONVERSION": "YES",
					"CLANG_WARN_DIRECT_OBJC_ISA_USAGE": "YES_ERROR",
					"CLANG_WARN_EMPTY_BODY": "YES",
					"CLANG_WARN_ENUM_CONVERSION": "YES",
					"CLANG_WARN_INFINITE_RECURSION": "YES",
					"CLANG_WARN_INT_CONVERSION": "YES",
					"CLANG_WARN_OBJC_ROOT_CLASS": "YES_ERROR",
					"CLANG_WARN_SUSPICIOUS_MOVE": "YES",
					"CLANG_WARN_UNREACHABLE_CODE": "YES",
					"CLANG_WARN__DUPLICATE_METHOD_MATCH": "YES",
					"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
					"COPY_PHASE_STRIP": "YES",
					"ENABLE_NS_ASSERTIONS": "NO",
					"ENABLE_STRICT_OBJC_MSGSEND": "YES",
					"GCC_C_LANGUAGE_STANDARD": "gnu99",
					"GCC_NO_COMMON_BLOCKS": "YES",
					"GCC_WARN_64_TO_32_BIT_CONVERSION": "YES",
					"GCC_WARN_ABOUT_RETURN_TYPE": "YES_ERROR",
					"GCC_WARN_UNDECLARED_SELECTOR": "YES",
					"GCC_WARN_UNINITIALIZED_AUTOS": "YES_AGGRESSIVE",
					"GCC_WARN_UNUSED_FUNCTION": "YES",
					"GCC_WARN_UNUSED_VARIABLE": "YES",
					"IPHONEOS_DEPLOYMENT_TARGET": "8.1",
					"MTL_ENABLE_DEBUG_INFO": "NO",
					"SDKROOT": "iphoneos",
					"TARGETED_DEVICE_FAMILY": "1,2",
					"VALIDATE_PRODUCT": "YES"
				}
			}
		]
	},
	"Targets": [
		{
			"Type": "PBXNativeTarget",
			"ID": "BA3CBE7419F7A93800CED4D5",
			"Name": "ios-simple-objc",
			"BuildConfigurationList": {
				"ID": "BA3CBE9B19F7A93900CED4D5",
				"DefaultConfigurationName": "Release",
				"BuildConfigurations": [
					{
						"ID": "BA3CBE9C19F7A93900CED4D5",
						"Name": "Debug",
						"BuildSettings": {
							"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
							"CODE_SIGN_IDENTITY": "iPhone Developer",
							"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
							"DEVELOPMENT_TEAM": "72SA8V3WYL",
							"INFOPLIST_FILE": "ios-simple-objc/Info.plist",
							"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks",
							"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
							"PRODUCT_NAME": "$(TARGET_NAME)",
							"PROVISIONING_PROFILE": "",
							"PROVISIONING_PROFILE_SPECIFIER": "BitriseBot-Wildcard"
						}
					},
					{
						"ID": "BA3CBE9D19F7A93900CED4D5",
						"Name": "Release",
						"BuildSettings": {
							"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
							"CODE_SIGN_IDENTITY": "iPhone Developer",
							"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
							"DEVELOPMENT_TEAM": "72SA8V3WYL",
							"INFOPLIST_FILE": "ios-simple-objc/Info.plist",
							"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks",
							"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
							"PRODUCT_NAME": "$(TARGET_NAME)",
							"PROVISIONING_PROFILE": "",
							"PROVISIONING_PROFILE_SPECIFIER": "BitriseBot-Wildcard"
						}
					}
				]
			},
			"Dependencies": null,
			"ProductReference": {
				"Path": "ios-simple-objc.app"
			},
			"ProductType": "com.apple.product-type.application"
		},
		{
			"Type": "PBXNativeTarget",
			"ID": "BA3CBE9019F7A93900CED4D5",
			"Name": "ios-simple-objcTests",
			"BuildConfigurationList": {
				"ID": "BA3CBE9E19F7A93900CED4D5",
				"DefaultConfigurationName": "Release",
				"BuildConfigurations": [
					{
						"ID": "BA3CBE9F19F7A93900CED4D5",
						"Name": "Debug",
						"BuildSettings": {
							"BUNDLE_LOADER": "$(TEST_HOST)",
							"FRAMEWORK_SEARCH_PATHS": [
								"$(SDKROOT)/Developer/Library/Frameworks",
								"$(inherited)"
							],
							"GCC_PREPROCESSOR_DEFINITIONS": [
								"DEBUG=1",
								"$(inherited)"
							],
							"INFOPLIST_FILE": "ios-simple-objcTests/Info.plist",
							"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks @loader_path/Frameworks",
							"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
							"PRODUCT_NAME": "$(TARGET_NAME)",
							"TEST_HOST": "$(BUILT_PRODUCTS_DIR)/ios-simple-objc.app/ios-simple-objc"
						}
					},
					{
						"ID": "BA3CBEA019F7A93900CED4D5",
						"Name": "Release",
						"BuildSettings": {
							"BUNDLE_LOADER": "$(TEST_HOST)",
							"FRAMEWORK_SEARCH_PATHS": [
								"$(SDKROOT)/Developer/Library/Frameworks",
								"$(inherited)"
							],
							"INFOPLIST_FILE": "ios-simple-objcTests/Info.plist",
							"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks @loader_path/Frameworks",
							"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
							"PRODUCT_NAME": "$(TARGET_NAME)",
							"TEST_HOST": "$(BUILT_PRODUCTS_DIR)/ios-simple-objc.app/ios-simple-objc"
						}
					}
				]
			},
			"Dependencies": [
				{
					"ID": "BA3CBE9319F7A93900CED4D5",
					"Target": {
						"Type": "PBXNativeTarget",
						"ID": "BA3CBE7419F7A93800CED4D5",
						"Name": "ios-simple-objc",
						"BuildConfigurationList": {
							"ID": "BA3CBE9B19F7A93900CED4D5",
							"DefaultConfigurationName": "Release",
							"BuildConfigurations": [
								{
									"ID": "BA3CBE9C19F7A93900CED4D5",
									"Name": "Debug",
									"BuildSettings": {
										"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
										"CODE_SIGN_IDENTITY": "iPhone Developer",
										"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
										"DEVELOPMENT_TEAM": "72SA8V3WYL",
										"INFOPLIST_FILE": "ios-simple-objc/Info.plist",
										"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks",
										"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
										"PRODUCT_NAME": "$(TARGET_NAME)",
										"PROVISIONING_PROFILE": "",
										"PROVISIONING_PROFILE_SPECIFIER": "BitriseBot-Wildcard"
									}
								},
								{
									"ID": "BA3CBE9D19F7A93900CED4D5",
									"Name": "Release",
									"BuildSettings": {
										"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
										"CODE_SIGN_IDENTITY": "iPhone Developer",
										"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
										"DEVELOPMENT_TEAM": "72SA8V3WYL",
										"INFOPLIST_FILE": "ios-simple-objc/Info.plist",
										"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks",
										"PRODUCT_BUNDLE_IDENTIFIER": "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
										"PRODUCT_NAME": "$(TARGET_NAME)",
										"PROVISIONING_PROFILE": "",
										"PROVISIONING_PROFILE_SPECIFIER": "BitriseBot-Wildcard"
									}
								}
							]
						},
						"Dependencies": null,
						"ProductReference": {
							"Path": "ios-simple-objc.app"
						},
						"ProductType": "com.apple.product-type.application"
					}
				}
			],
			"ProductReference": {
				"Path": "ios-simple-objcTests.xctest"
			},
			"ProductType": "com.apple.product-type.bundle.unit-test"
		}
	],
	"Attributes": {
		"TargetAttributes": {
			"BA3CBE7419F7A93800CED4D5": {
				"CreatedOnToolsVersion": "6.1",
				"DevelopmentTeam": "72SA8V3WYL",
				"ProvisioningStyle": "Manual"
			},
			"BA3CBE9019F7A93900CED4D5": {
				"CreatedOnToolsVersion": "6.1",
				"TestTargetID": "BA3CBE7419F7A93800CED4D5"
			}
		}
	}
}`
