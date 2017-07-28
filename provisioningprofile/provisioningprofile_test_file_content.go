package provisioningprofile

const developmentProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>Bitrise Test</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>9NS44DLTN7</string>
	</array>
	<key>CreationDate</key>
	<date>2016-09-22T11:28:46Z</date>
	<key>Platform</key>
	<array>
		<string>iOS</string>
	</array>
	<key>DeveloperCertificates</key>
	<array>
		<data></data>
	</array>
	<key>Entitlements</key>
	<dict>
		<key>keychain-access-groups</key>
		<array>
			<string>9NS44DLTN7.*</string>
		</array>
		<key>get-task-allow</key>
		<true/>
		<key>application-identifier</key>
		<string>9NS44DLTN7.*</string>
		<key>com.apple.developer.team-identifier</key>
		<string>9NS44DLTN7</string>
	</dict>
	<key>ExpirationDate</key>
	<date>2017-09-22T11:28:46Z</date>
	<key>Name</key>
	<string>Bitrise Test Development</string>
	<key>ProvisionedDevices</key>
	<array>
		<string>b13813075ad9b298cb9a9f28555c49573d8bc322</string>
	</array>
	<key>TeamIdentifier</key>
	<array>
		<string>9NS44DLTN7</string>
	</array>
	<key>TeamName</key>
	<string>Some Dude</string>
	<key>TimeToLive</key>
	<integer>365</integer>
	<key>UUID</key>
	<string>4b617a5f-e31e-4edc-9460-718a5abacd05</string>
	<key>Version</key>
	<integer>1</integer>
</dict>`

const appStoreProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>Bitrise Test</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>9NS44DLTN7</string>
	</array>
	<key>CreationDate</key>
	<date>2016-09-22T11:29:12Z</date>
	<key>Platform</key>
	<array>
		<string>iOS</string>
	</array>
	<key>DeveloperCertificates</key>
	<array>
		<data></data>
	</array>
	<key>Entitlements</key>
	<dict>
		<key>keychain-access-groups</key>
		<array>
			<string>9NS44DLTN7.*</string>
		</array>
		<key>get-task-allow</key>
		<false/>
		<key>application-identifier</key>
		<string>9NS44DLTN7.*</string>
		<key>com.apple.developer.team-identifier</key>
		<string>9NS44DLTN7</string>
		<key>beta-reports-active</key>
		<true/>
	</dict>
	<key>ExpirationDate</key>
	<date>2017-09-21T13:20:06Z</date>
	<key>Name</key>
	<string>Bitrise Test App Store</string>
	<key>TeamIdentifier</key>
	<array>
		<string>9NS44DLTN7</string>
	</array>
	<key>TeamName</key>
	<string>Some Dude</string>
	<key>TimeToLive</key>
	<integer>364</integer>
	<key>UUID</key>
	<string>a60668dd-191a-4770-8b1e-b453b87aa60b</string>
	<key>Version</key>
	<integer>1</integer>
</dict>`

const adHocProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>Bitrise Test</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>9NS44DLTN7</string>
	</array>
	<key>CreationDate</key>
	<date>2016-09-22T11:29:38Z</date>
	<key>Platform</key>
	<array>
		<string>iOS</string>
	</array>
	<key>DeveloperCertificates</key>
	<array>
		<data></data>
	</array>
	<key>Entitlements</key>
	<dict>
		<key>keychain-access-groups</key>
		<array>
			<string>9NS44DLTN7.*</string>
		</array>
		<key>get-task-allow</key>
		<false/>
		<key>application-identifier</key>
		<string>9NS44DLTN7.*</string>
		<key>com.apple.developer.team-identifier</key>
		<string>9NS44DLTN7</string>
	</dict>
	<key>ExpirationDate</key>
	<date>2017-09-21T13:20:06Z</date>
	<key>Name</key>
	<string>Bitrise Test Ad Hoc</string>
	<key>ProvisionedDevices</key>
	<array>
		<string>b13813075ad9b298cb9a9f28555c49573d8bc322</string>
	</array>
	<key>TeamIdentifier</key>
	<array>
		<string>9NS44DLTN7</string>
	</array>
	<key>TeamName</key>
	<string>Some Dude</string>
	<key>TimeToLive</key>
	<integer>364</integer>
	<key>UUID</key>
	<string>26668300-5743-46a1-8e00-7023e2e35c7d</string>
	<key>Version</key>
	<integer>1</integer>
</dict>`

const enterpriseProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>PaintSpeciPad</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>PF3BP78LQ8</string>
	</array>
	<key>CreationDate</key>
	<date>2015-10-05T13:32:46Z</date>
	<key>Platform</key>
	<array>
		<string>iOS</string>
	</array>
	<key>DeveloperCertificates</key>
	<array>
		<data></data>
	</array>
	<key>Entitlements</key>
	<dict>
		<key>keychain-access-groups</key>
		<array>
			<string>PF3BP78LQ8.*</string>
		</array>
		<key>get-task-allow</key>
		<false/>
		<key>application-identifier</key>
		<string>PF3BP78LQ8.com.akzonobel.PaintSpeciPad</string>
		<key>com.apple.developer.team-identifier</key>
		<string>PF3BP78LQ8</string>

	</dict>
	<key>ExpirationDate</key>
	<date>2016-10-04T13:32:46Z</date>
	<key>Name</key>
	<string>PaintSpeciPadDistProf</string>
	<key>ProvisionsAllDevices</key>
	<true/>
	<key>TeamIdentifier</key>
	<array>
		<string>PF3BP78LQ8</string>
	</array>
	<key>TeamName</key>
	<string>Akzo Nobel Decorative Coatings B.V.</string>
	<key>TimeToLive</key>
	<integer>365</integer>
	<key>UUID</key>
	<string>8d6caa15-ac49-48f9-9bd3-ce9244add6a0</string>
	<key>Version</key>
	<integer>1</integer>
</dict>`

const buildSettingsOut = `2017-07-28 08:54:23.860 xcodebuild[8023:63576] CoreSimulator detected Xcode.app relocation or CoreSimulatorService version change.  Framework path (/Library/Developer/PrivateFrameworks/CoreSimulator.framework) and version (494.4) does not match existing job path (/Applications/Xcode.app/Contents/Developer/Library/PrivateFrameworks/CoreSimulator.framework/Versions/A/XPCServices/com.apple.CoreSimulator.CoreSimulatorService.xpc) and version (375.21).  Attempting to remove the stale service in order to add the expected version.
Build settings for action build and target ios-simple-objc:
    ACTION = build
    AD_HOC_CODE_SIGNING_ALLOWED = NO
    ALTERNATE_GROUP = staff
    ALTERNATE_MODE = u+w,go-w,a+rX
    ALTERNATE_OWNER = bitrise
    ALWAYS_EMBED_SWIFT_STANDARD_LIBRARIES = NO
    ALWAYS_SEARCH_USER_PATHS = NO
    ALWAYS_USE_SEPARATE_HEADERMAPS = NO
    APPLE_INTERNAL_DEVELOPER_DIR = /AppleInternal/Developer
    APPLE_INTERNAL_DIR = /AppleInternal
    APPLE_INTERNAL_DOCUMENTATION_DIR = /AppleInternal/Documentation
    APPLE_INTERNAL_LIBRARY_DIR = /AppleInternal/Library
    APPLE_INTERNAL_TOOLS = /AppleInternal/Developer/Tools
    APPLICATION_EXTENSION_API_ONLY = NO
    APPLY_RULES_IN_COPY_FILES = NO
    ARCHS = armv7 arm64
    ARCHS_STANDARD = armv7 arm64
    ARCHS_STANDARD_32_64_BIT = armv7 arm64
    ARCHS_STANDARD_32_BIT = armv7
    ARCHS_STANDARD_64_BIT = arm64
    ARCHS_STANDARD_INCLUDING_64_BIT = armv7 arm64
    ARCHS_UNIVERSAL_IPHONE_OS = armv7 arm64
    ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon
    AVAILABLE_PLATFORMS = appletvos appletvsimulator iphoneos iphonesimulator macosx watchos watchsimulator
    BITCODE_GENERATION_MODE = marker
    BUILD_ACTIVE_RESOURCES_ONLY = NO
    BUILD_COMPONENTS = headers build
    BUILD_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products
    BUILD_ROOT = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products
    BUILD_STYLE =
    BUILD_VARIANTS = normal
    BUILT_PRODUCTS_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos
    CACHE_ROOT = /var/folders/77/53svxlss13jg3f0nz_jc8vch0000gn/C/com.apple.DeveloperTools/9.0-9M189t/Xcode
    CCHROOT = /var/folders/77/53svxlss13jg3f0nz_jc8vch0000gn/C/com.apple.DeveloperTools/9.0-9M189t/Xcode
    CHMOD = /bin/chmod
    CHOWN = /usr/sbin/chown
    CLANG_CXX_LANGUAGE_STANDARD = gnu++0x
    CLANG_CXX_LIBRARY = libc++
    CLANG_ENABLE_MODULES = YES
    CLANG_ENABLE_OBJC_ARC = YES
    CLANG_WARN_BOOL_CONVERSION = YES
    CLANG_WARN_CONSTANT_CONVERSION = YES
    CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR
    CLANG_WARN_EMPTY_BODY = YES
    CLANG_WARN_ENUM_CONVERSION = YES
    CLANG_WARN_INFINITE_RECURSION = YES
    CLANG_WARN_INT_CONVERSION = YES
    CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR
    CLANG_WARN_SUSPICIOUS_MOVE = YES
    CLANG_WARN_UNREACHABLE_CODE = YES
    CLANG_WARN__DUPLICATE_METHOD_MATCH = YES
    CLASS_FILE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/JavaClasses
    CLEAN_PRECOMPS = YES
    CLONE_HEADERS = NO
    CODESIGNING_FOLDER_PATH = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos/ios-simple-objc.app
    CODE_SIGNING_ALLOWED = YES
    CODE_SIGNING_REQUIRED = YES
    CODE_SIGN_CONTEXT_CLASS = XCiPhoneOSCodeSignContext
    CODE_SIGN_IDENTITY = iPhone Developer
    COLOR_DIAGNOSTICS = YES
    COMBINE_HIDPI_IMAGES = NO
    COMPOSITE_SDK_DIRS = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/CompositeSDKs
    COMPRESS_PNG_FILES = YES
    CONFIGURATION = Release
    CONFIGURATION_BUILD_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos
    CONFIGURATION_TEMP_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos
    CONTENTS_FOLDER_PATH = ios-simple-objc.app
    COPYING_PRESERVES_HFS_DATA = NO
    COPY_HEADERS_RUN_UNIFDEF = NO
    COPY_PHASE_STRIP = YES
    COPY_RESOURCES_FROM_STATIC_FRAMEWORKS = YES
    CORRESPONDING_SIMULATOR_PLATFORM_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneSimulator.platform
    CORRESPONDING_SIMULATOR_PLATFORM_NAME = iphonesimulator
    CORRESPONDING_SIMULATOR_SDK_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/SDKs/iPhoneSimulator11.0.sdk
    CORRESPONDING_SIMULATOR_SDK_NAME = iphonesimulator11.0
    CP = /bin/cp
    CREATE_INFOPLIST_SECTION_IN_BINARY = NO
    CURRENT_ARCH = arm64
    CURRENT_VARIANT = normal
    DEAD_CODE_STRIPPING = YES
    DEBUGGING_SYMBOLS = YES
    DEBUG_INFORMATION_FORMAT = dwarf-with-dsym
    DEFAULT_COMPILER = com.apple.compilers.llvm.clang.1_0
    DEFAULT_KEXT_INSTALL_PATH = /System/Library/Extensions
    DEFINES_MODULE = NO
    DEPLOYMENT_LOCATION = NO
    DEPLOYMENT_POSTPROCESSING = NO
    DEPLOYMENT_TARGET_CLANG_ENV_NAME = IPHONEOS_DEPLOYMENT_TARGET
    DEPLOYMENT_TARGET_CLANG_FLAG_NAME = miphoneos-version-min
    DEPLOYMENT_TARGET_CLANG_FLAG_PREFIX = -miphoneos-version-min=
    DEPLOYMENT_TARGET_SETTING_NAME = IPHONEOS_DEPLOYMENT_TARGET
    DEPLOYMENT_TARGET_SUGGESTED_VALUES = 8.0 8.1 8.2 8.3 8.4 9.0 9.1 9.2 9.3 10.0 10.1 10.2 10.3 11.0
    DERIVED_FILES_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/DerivedSources
    DERIVED_FILE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/DerivedSources
    DERIVED_SOURCES_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/DerivedSources
    DEVELOPER_APPLICATIONS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications
    DEVELOPER_BIN_DIR = /Applications/Xcode-beta.app/Contents/Developer/usr/bin
    DEVELOPER_DIR = /Applications/Xcode-beta.app/Contents/Developer
    DEVELOPER_FRAMEWORKS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Library/Frameworks
    DEVELOPER_FRAMEWORKS_DIR_QUOTED = /Applications/Xcode-beta.app/Contents/Developer/Library/Frameworks
    DEVELOPER_LIBRARY_DIR = /Applications/Xcode-beta.app/Contents/Developer/Library
    DEVELOPER_SDK_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs
    DEVELOPER_TOOLS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Tools
    DEVELOPER_USR_DIR = /Applications/Xcode-beta.app/Contents/Developer/usr
    DEVELOPMENT_LANGUAGE = English
    DEVELOPMENT_TEAM = 72SA8V3WYL
    DOCUMENTATION_FOLDER_PATH = ios-simple-objc.app/English.lproj/Documentation
    DO_HEADER_SCANNING_IN_JAM = NO
    DSTROOT = /tmp/ios-simple-objc.dst
    DT_TOOLCHAIN_DIR = /Applications/Xcode-beta.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain
    DWARF_DSYM_FILE_NAME = ios-simple-objc.app.dSYM
    DWARF_DSYM_FILE_SHOULD_ACCOMPANY_PRODUCT = NO
    DWARF_DSYM_FOLDER_PATH = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos
    EFFECTIVE_PLATFORM_NAME = -iphoneos
    EMBEDDED_CONTENT_CONTAINS_SWIFT = NO
    EMBEDDED_PROFILE_NAME = embedded.mobileprovision
    EMBED_ASSET_PACKS_IN_PRODUCT_BUNDLE = NO
    ENABLE_BITCODE = YES
    ENABLE_DEFAULT_HEADER_SEARCH_PATHS = YES
    ENABLE_HEADER_DEPENDENCIES = YES
    ENABLE_NS_ASSERTIONS = NO
    ENABLE_ON_DEMAND_RESOURCES = YES
    ENABLE_STRICT_OBJC_MSGSEND = YES
    ENABLE_TESTABILITY = NO
    ENTITLEMENTS_ALLOWED = YES
    ENTITLEMENTS_REQUIRED = YES
    EXCLUDED_INSTALLSRC_SUBDIRECTORY_PATTERNS = .DS_Store .svn .git .hg CVS
    EXCLUDED_RECURSIVE_SEARCH_PATH_SUBDIRECTORIES = *.nib *.lproj *.framework *.gch *.xcode* *.xcassets (*) .DS_Store CVS .svn .git .hg *.pbproj *.pbxproj
    EXECUTABLES_FOLDER_PATH = ios-simple-objc.app/Executables
    EXECUTABLE_FOLDER_PATH = ios-simple-objc.app
    EXECUTABLE_NAME = ios-simple-objc
    EXECUTABLE_PATH = ios-simple-objc.app/ios-simple-objc
    EXPANDED_CODE_SIGN_IDENTITY =
    EXPANDED_CODE_SIGN_IDENTITY_NAME =
    EXPANDED_PROVISIONING_PROFILE =
    FILE_LIST = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/Objects/LinkFileList
    FIXED_FILES_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/FixedFiles
    FRAMEWORKS_FOLDER_PATH = ios-simple-objc.app/Frameworks
    FRAMEWORK_FLAG_PREFIX = -framework
    FRAMEWORK_VERSION = A
    FULL_PRODUCT_NAME = ios-simple-objc.app
    GCC3_VERSION = 3.3
    GCC_C_LANGUAGE_STANDARD = gnu99
    GCC_INLINES_ARE_PRIVATE_EXTERN = YES
    GCC_NO_COMMON_BLOCKS = YES
    GCC_PFE_FILE_C_DIALECTS = c objective-c c++ objective-c++
    GCC_SYMBOLS_PRIVATE_EXTERN = YES
    GCC_THUMB_SUPPORT = YES
    GCC_TREAT_WARNINGS_AS_ERRORS = NO
    GCC_VERSION = com.apple.compilers.llvm.clang.1_0
    GCC_VERSION_IDENTIFIER = com_apple_compilers_llvm_clang_1_0
    GCC_WARN_64_TO_32_BIT_CONVERSION = YES
    GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR
    GCC_WARN_UNDECLARED_SELECTOR = YES
    GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE
    GCC_WARN_UNUSED_FUNCTION = YES
    GCC_WARN_UNUSED_VARIABLE = YES
    GENERATE_MASTER_OBJECT_FILE = NO
    GENERATE_PKGINFO_FILE = YES
    GENERATE_PROFILING_CODE = NO
    GENERATE_TEXT_BASED_STUBS = NO
    GID = 20
    GROUP = staff
    HEADERMAP_INCLUDES_FLAT_ENTRIES_FOR_TARGET_BEING_BUILT = YES
    HEADERMAP_INCLUDES_FRAMEWORK_ENTRIES_FOR_ALL_PRODUCT_TYPES = YES
    HEADERMAP_INCLUDES_NONPUBLIC_NONPRIVATE_HEADERS = YES
    HEADERMAP_INCLUDES_PROJECT_HEADERS = YES
    HEADERMAP_USES_FRAMEWORK_PREFIX_ENTRIES = YES
    HEADERMAP_USES_VFS = NO
    HIDE_BITCODE_SYMBOLS = YES
    HOME = /Users/bitrise
    ICONV = /usr/bin/iconv
    INFOPLIST_EXPAND_BUILD_SETTINGS = YES
    INFOPLIST_FILE = ios-simple-objc/Info.plist
    INFOPLIST_OUTPUT_FORMAT = binary
    INFOPLIST_PATH = ios-simple-objc.app/Info.plist
    INFOPLIST_PREPROCESS = NO
    INFOSTRINGS_PATH = ios-simple-objc.app/English.lproj/InfoPlist.strings
    INLINE_PRIVATE_FRAMEWORKS = NO
    INSTALLHDRS_COPY_PHASE = NO
    INSTALLHDRS_SCRIPT_PHASE = NO
    INSTALL_DIR = /tmp/ios-simple-objc.dst/Applications
    INSTALL_GROUP = staff
    INSTALL_MODE_FLAG = u+w,go-w,a+rX
    INSTALL_OWNER = bitrise
    INSTALL_PATH = /Applications
    INSTALL_ROOT = /tmp/ios-simple-objc.dst
    IPHONEOS_DEPLOYMENT_TARGET = 8.1
    JAVAC_DEFAULT_FLAGS = -J-Xms64m -J-XX:NewSize=4M -J-Dfile.encoding=UTF8
    JAVA_APP_STUB = /System/Library/Frameworks/JavaVM.framework/Resources/MacOS/JavaApplicationStub
    JAVA_ARCHIVE_CLASSES = YES
    JAVA_ARCHIVE_TYPE = JAR
    JAVA_COMPILER = /usr/bin/javac
    JAVA_FOLDER_PATH = ios-simple-objc.app/Java
    JAVA_FRAMEWORK_RESOURCES_DIRS = Resources
    JAVA_JAR_FLAGS = cv
    JAVA_SOURCE_SUBDIR = .
    JAVA_USE_DEPENDENCIES = YES
    JAVA_ZIP_FLAGS = -urg
    JIKES_DEFAULT_FLAGS = +E +OLDCSO
    KASAN_DEFAULT_CFLAGS = -DKASAN=1 -fsanitize=address -mllvm -asan-globals-live-support -mllvm -asan-force-dynamic-shadow
    KEEP_PRIVATE_EXTERNS = NO
    LD_DEPENDENCY_INFO_FILE = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/Objects-normal/arm64/ios-simple-objc_dependency_info.dat
    LD_GENERATE_MAP_FILE = NO
    LD_MAP_FILE_PATH = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/ios-simple-objc-LinkMap-normal-arm64.txt
    LD_NO_PIE = NO
    LD_QUOTE_LINKER_ARGUMENTS_FOR_COMPILER_DRIVER = YES
    LD_RUNPATH_SEARCH_PATHS =  @executable_path/Frameworks
    LEGACY_DEVELOPER_DIR = /Applications/Xcode-beta.app/Contents/PlugIns/Xcode3Core.ideplugin/Contents/SharedSupport/Developer
    LEX = lex
    LIBRARY_FLAG_NOSPACE = YES
    LIBRARY_FLAG_PREFIX = -l
    LIBRARY_KEXT_INSTALL_PATH = /Library/Extensions
    LINKER_DISPLAYS_MANGLED_NAMES = NO
    LINK_FILE_LIST_normal_arm64 =
    LINK_FILE_LIST_normal_armv7 =
    LINK_WITH_STANDARD_LIBRARIES = YES
    LOCALIZABLE_CONTENT_DIR =
    LOCALIZED_RESOURCES_FOLDER_PATH = ios-simple-objc.app/English.lproj
    LOCAL_ADMIN_APPS_DIR = /Applications/Utilities
    LOCAL_APPS_DIR = /Applications
    LOCAL_DEVELOPER_DIR = /Library/Developer
    LOCAL_LIBRARY_DIR = /Library
    LOCROOT =
    LOCSYMROOT =
    MACH_O_TYPE = mh_execute
    MAC_OS_X_PRODUCT_BUILD_VERSION = 16G29
    MAC_OS_X_VERSION_ACTUAL = 101206
    MAC_OS_X_VERSION_MAJOR = 101200
    MAC_OS_X_VERSION_MINOR = 1206
    METAL_LIBRARY_FILE_BASE = default
    METAL_LIBRARY_OUTPUT_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos/ios-simple-objc.app
    MODULE_CACHE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ModuleCache
    MTL_ENABLE_DEBUG_INFO = NO
    NATIVE_ARCH = armv7
    NATIVE_ARCH_32_BIT = i386
    NATIVE_ARCH_64_BIT = x86_64
    NATIVE_ARCH_ACTUAL = x86_64
    NO_COMMON = YES
    OBJECT_FILE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/Objects
    OBJECT_FILE_DIR_normal = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/Objects-normal
    OBJROOT = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex
    ONLY_ACTIVE_ARCH = NO
    OS = MACOS
    OSAC = /usr/bin/osacompile
    PACKAGE_TYPE = com.apple.package-type.wrapper.application
    PASCAL_STRINGS = YES
    PATH = /Applications/Xcode-beta.app/Contents/Developer/usr/bin:/Users/bitrise/.fastlane/bin:/Users/bitrise/.rbenv/shims:/Library/Frameworks/Mono.framework/Versions/Current/Commands:/Users/bitrise/Develop/go/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/opt/X11/bin:/Library/Frameworks/Mono.framework/Versions/Current/Commands
    PATH_PREFIXES_EXCLUDED_FROM_HEADER_DEPENDENCIES = /usr/include /usr/local/include /System/Library/Frameworks /System/Library/PrivateFrameworks /Applications/Xcode-beta.app/Contents/Developer/Headers /Applications/Xcode-beta.app/Contents/Developer/SDKs /Applications/Xcode-beta.app/Contents/Developer/Platforms
    PBDEVELOPMENTPLIST_PATH = ios-simple-objc.app/pbdevelopment.plist
    PFE_FILE_C_DIALECTS = objective-c
    PKGINFO_FILE_PATH = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/PkgInfo
    PKGINFO_PATH = ios-simple-objc.app/PkgInfo
    PLATFORM_DEVELOPER_APPLICATIONS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/Applications
    PLATFORM_DEVELOPER_BIN_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/usr/bin
    PLATFORM_DEVELOPER_LIBRARY_DIR = /Applications/Xcode-beta.app/Contents/PlugIns/Xcode3Core.ideplugin/Contents/SharedSupport/Developer/Library
    PLATFORM_DEVELOPER_SDK_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs
    PLATFORM_DEVELOPER_TOOLS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/Tools
    PLATFORM_DEVELOPER_USR_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/usr
    PLATFORM_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform
    PLATFORM_DISPLAY_NAME = iOS
    PLATFORM_NAME = iphoneos
    PLATFORM_PREFERRED_ARCH = arm64
    PLATFORM_PRODUCT_BUILD_VERSION = 15A5327g
    PLIST_FILE_OUTPUT_FORMAT = binary
    PLUGINS_FOLDER_PATH = ios-simple-objc.app/PlugIns
    PRECOMPS_INCLUDE_HEADERS_FROM_BUILT_PRODUCTS_DIR = YES
    PRECOMP_DESTINATION_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/PrefixHeaders
    PRESERVE_DEAD_CODE_INITS_AND_TERMS = NO
    PRIVATE_HEADERS_FOLDER_PATH = ios-simple-objc.app/PrivateHeaders
    PRODUCT_BUNDLE_IDENTIFIER = Bitrise.ios-simple-objc
    PRODUCT_MODULE_NAME = ios_simple_objc
    PRODUCT_NAME = ios-simple-objc
    PRODUCT_SETTINGS_PATH = /Users/bitrise/Develop/go/src/github.com/bitrise-io/steps-xcode-archive/_tmp/ios-simple-objc/ios-simple-objc/Info.plist
    PRODUCT_TYPE = com.apple.product-type.application
    PROFILING_CODE = NO
    PROJECT = ios-simple-objc
    PROJECT_DERIVED_FILE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/DerivedSources
    PROJECT_DIR = /Users/bitrise/Develop/go/src/github.com/bitrise-io/steps-xcode-archive/_tmp/ios-simple-objc
    PROJECT_FILE_PATH = /Users/bitrise/Develop/go/src/github.com/bitrise-io/steps-xcode-archive/_tmp/ios-simple-objc/ios-simple-objc.xcodeproj
    PROJECT_NAME = ios-simple-objc
    PROJECT_TEMP_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build
    PROJECT_TEMP_ROOT = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex
    PROVISIONING_PROFILE_REQUIRED = YES
    PROVISIONING_PROFILE_SPECIFIER = BitriseBot-Wildcard
    PUBLIC_HEADERS_FOLDER_PATH = ios-simple-objc.app/Headers
    RECURSIVE_SEARCH_PATHS_FOLLOW_SYMLINKS = YES
    REMOVE_CVS_FROM_RESOURCES = YES
    REMOVE_GIT_FROM_RESOURCES = YES
    REMOVE_HEADERS_FROM_EMBEDDED_BUNDLES = YES
    REMOVE_HG_FROM_RESOURCES = YES
    REMOVE_SVN_FROM_RESOURCES = YES
    RESOURCE_RULES_REQUIRED = YES
    REZ_COLLECTOR_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/ResourceManagerResources
    REZ_OBJECTS_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build/ResourceManagerResources/Objects
    SCAN_ALL_SOURCE_FILES_FOR_INCLUDES = NO
    SCRIPTS_FOLDER_PATH = ios-simple-objc.app/Scripts
    SDKROOT = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS11.0.sdk
    SDK_DIR = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS11.0.sdk
    SDK_DIR_iphoneos11_0 = /Applications/Xcode-beta.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS11.0.sdk
    SDK_NAME = iphoneos11.0
    SDK_NAMES = iphoneos11.0
    SDK_PRODUCT_BUILD_VERSION = 15A5327g
    SDK_VERSION = 11.0
    SDK_VERSION_ACTUAL = 110000
    SDK_VERSION_MAJOR = 110000
    SDK_VERSION_MINOR = 000
    SED = /usr/bin/sed
    SEPARATE_STRIP = NO
    SEPARATE_SYMBOL_EDIT = NO
    SET_DIR_MODE_OWNER_GROUP = YES
    SET_FILE_MODE_OWNER_GROUP = NO
    SHALLOW_BUNDLE = YES
    SHARED_DERIVED_FILE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos/DerivedSources
    SHARED_FRAMEWORKS_FOLDER_PATH = ios-simple-objc.app/SharedFrameworks
    SHARED_PRECOMPS_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/PrecompiledHeaders
    SHARED_SUPPORT_FOLDER_PATH = ios-simple-objc.app/SharedSupport
    SKIP_INSTALL = NO
    SOURCE_ROOT = /Users/bitrise/Develop/go/src/github.com/bitrise-io/steps-xcode-archive/_tmp/ios-simple-objc
    SRCROOT = /Users/bitrise/Develop/go/src/github.com/bitrise-io/steps-xcode-archive/_tmp/ios-simple-objc
    STRINGS_FILE_OUTPUT_ENCODING = binary
    STRIP_BITCODE_FROM_COPIED_FILES = YES
    STRIP_INSTALLED_PRODUCT = YES
    STRIP_STYLE = all
    STRIP_SWIFT_SYMBOLS = YES
    SUPPORTED_DEVICE_FAMILIES = 1,2
    SUPPORTED_PLATFORMS = iphonesimulator iphoneos
    SUPPORTS_TEXT_BASED_API = NO
    SWIFT_PLATFORM_TARGET_PREFIX = ios
    SYMROOT = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products
    SYSTEM_ADMIN_APPS_DIR = /Applications/Utilities
    SYSTEM_APPS_DIR = /Applications
    SYSTEM_CORE_SERVICES_DIR = /System/Library/CoreServices
    SYSTEM_DEMOS_DIR = /Applications/Extras
    SYSTEM_DEVELOPER_APPS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications
    SYSTEM_DEVELOPER_BIN_DIR = /Applications/Xcode-beta.app/Contents/Developer/usr/bin
    SYSTEM_DEVELOPER_DEMOS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications/Utilities/Built Examples
    SYSTEM_DEVELOPER_DIR = /Applications/Xcode-beta.app/Contents/Developer
    SYSTEM_DEVELOPER_DOC_DIR = /Applications/Xcode-beta.app/Contents/Developer/ADC Reference Library
    SYSTEM_DEVELOPER_GRAPHICS_TOOLS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications/Graphics Tools
    SYSTEM_DEVELOPER_JAVA_TOOLS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications/Java Tools
    SYSTEM_DEVELOPER_PERFORMANCE_TOOLS_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications/Performance Tools
    SYSTEM_DEVELOPER_RELEASENOTES_DIR = /Applications/Xcode-beta.app/Contents/Developer/ADC Reference Library/releasenotes
    SYSTEM_DEVELOPER_TOOLS = /Applications/Xcode-beta.app/Contents/Developer/Tools
    SYSTEM_DEVELOPER_TOOLS_DOC_DIR = /Applications/Xcode-beta.app/Contents/Developer/ADC Reference Library/documentation/DeveloperTools
    SYSTEM_DEVELOPER_TOOLS_RELEASENOTES_DIR = /Applications/Xcode-beta.app/Contents/Developer/ADC Reference Library/releasenotes/DeveloperTools
    SYSTEM_DEVELOPER_USR_DIR = /Applications/Xcode-beta.app/Contents/Developer/usr
    SYSTEM_DEVELOPER_UTILITIES_DIR = /Applications/Xcode-beta.app/Contents/Developer/Applications/Utilities
    SYSTEM_DOCUMENTATION_DIR = /Library/Documentation
    SYSTEM_KEXT_INSTALL_PATH = /System/Library/Extensions
    SYSTEM_LIBRARY_DIR = /System/Library
    TAPI_VERIFY_MODE = ErrorsOnly
    TARGETED_DEVICE_FAMILY = 1,2
    TARGETNAME = ios-simple-objc
    TARGET_BUILD_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Products/Release-iphoneos
    TARGET_NAME = ios-simple-objc
    TARGET_TEMP_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build
    TEMP_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build
    TEMP_FILES_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build
    TEMP_FILE_DIR = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex/ios-simple-objc.build/Release-iphoneos/ios-simple-objc.build
    TEMP_ROOT = /Users/bitrise/Library/Developer/Xcode/DerivedData/ios-simple-objc-bedswnqfbmnvzaghzbrixprblkyc/Build/Intermediates.noindex
    TOOLCHAIN_DIR = /Applications/Xcode-beta.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain
    TREAT_MISSING_BASELINES_AS_TEST_FAILURES = NO
    UID = 501
    UNLOCALIZED_RESOURCES_FOLDER_PATH = ios-simple-objc.app
    UNSTRIPPED_PRODUCT = NO
    USER = bitrise
    USER_APPS_DIR = /Users/bitrise/Applications
    USER_LIBRARY_DIR = /Users/bitrise/Library
    USE_DYNAMIC_NO_PIC = YES
    USE_HEADERMAP = YES
    USE_HEADER_SYMLINKS = NO
    VALIDATE_PRODUCT = YES
    VALID_ARCHS = arm64 armv7 armv7s
    VERBOSE_PBXCP = NO
    VERSIONPLIST_PATH = ios-simple-objc.app/version.plist
    VERSION_INFO_BUILDER = bitrise
    VERSION_INFO_FILE = ios-simple-objc_vers.c
    VERSION_INFO_STRING = "@(#)PROGRAM:ios-simple-objc  PROJECT:ios-simple-objc-"
    WRAPPER_EXTENSION = app
    WRAPPER_NAME = ios-simple-objc.app
    WRAPPER_SUFFIX = .app
    WRAP_ASSET_PACKS_IN_SEPARATE_DIRECTORIES = NO
    XCODE_APP_SUPPORT_DIR = /Applications/Xcode-beta.app/Contents/Developer/Library/Xcode
    XCODE_PRODUCT_BUILD_VERSION = 9M189t
    XCODE_VERSION_ACTUAL = 0900
    XCODE_VERSION_MAJOR = 0900
    XCODE_VERSION_MINOR = 0900
    XPCSERVICES_FOLDER_PATH = ios-simple-objc.app/XPCServices
    YACC = yacc
    arch = arm64
    diagnostic_message_length = 155
    variant = normal

`
