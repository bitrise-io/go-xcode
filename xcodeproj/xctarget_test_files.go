package xcodeproj

const (
	onlyApp = `
	// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 50;
	objects = {

/* Begin PBXNativeTarget section */
		64E1835F2588FD3C00D666BF /* test */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 645C5A6B2588F3F7004E3C82 /* Build configuration list for PBXNativeTarget "test" */;
			buildPhases = (
				645C5A532588F3F7004E3C82 /* Sources */,
				645C5A542588F3F7004E3C82 /* Frameworks */,
				645C5A552588F3F7004E3C82 /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = test;
			productName = test;
			productReference = 64E1835F2588FD3C00D666BF /* test.app */;
			productType = "com.apple.product-type.application";
		};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
		645C5A4F2588F3F7004E3C82 /* Project object */ = {
			isa = PBXProject;
			attributes = {
				LastSwiftUpdateCheck = 1220;
				LastUpgradeCheck = 1220;
				TargetAttributes = {
					645C5A562588F3F7004E3C82 = {
						CreatedOnToolsVersion = 12.2;
					};
				};
			};
			buildConfigurationList = 645C5A522588F3F7004E3C82 /* Build configuration list for PBXProject "test" */;
			compatibilityVersion = "Xcode 9.3";
			developmentRegion = en;
			hasScannedForEncodings = 0;
			knownRegions = (
				en,
				Base,
			);
			mainGroup = 645C5A4E2588F3F7004E3C82;
			productRefGroup = 645C5A582588F3F7004E3C82 /* Products */;
			projectDirPath = "";
			projectRoot = "";
			targets = (
				645C5A562588F3F7004E3C82 /* test */,
			);
		};
/* End PBXProject section */
	};
	rootObject = 645C5A4F2588F3F7004E3C82 /* Project object */;
}
`

	appWithAppClip = `
// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 50;
	objects = {

/* Begin PBXNativeTarget section */
		64E1835F2588FD3C00D666BF /* test */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E1838A2588FD3D00D666BF /* Build configuration list for PBXNativeTarget "test" */;
			buildPhases = (
				64E1835C2588FD3C00D666BF /* Sources */,
				64E1835D2588FD3C00D666BF /* Frameworks */,
				64E1835E2588FD3C00D666BF /* Resources */,
				64E183CA2588FD5F00D666BF /* Embed App Clips */,
				64E183F22588FD9E00D666BF /* Embed App Extensions */,
			);
			buildRules = (
			);
			dependencies = (
				64E183C52588FD5F00D666BF /* PBXTargetDependency */,
			);
			name = test;
			productName = test;
			productReference = 64E183602588FD3C00D666BF /* test.app */;
			productType = "com.apple.product-type.application";
		};
		64E1839B2588FD5E00D666BF /* clip */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E183C72588FD5F00D666BF /* Build configuration list for PBXNativeTarget "clip" */;
			buildPhases = (
				64E183982588FD5E00D666BF /* Sources */,
				64E183992588FD5E00D666BF /* Frameworks */,
				64E1839A2588FD5E00D666BF /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = clip;
			productName = clip;
			productReference = 64E1839C2588FD5E00D666BF /* clip.app */;
			productType = "com.apple.product-type.application.on-demand-install-capable";
		};
/* End PBXNativeTarget section */

/* Begin PBXTargetDependency section */
		64E183C52588FD5F00D666BF /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 64E1839B2588FD5E00D666BF /* clip */;
			targetProxy = 64E183C42588FD5F00D666BF /* PBXContainerItemProxy */;
		};
/* End PBXTargetDependency section */
	};
	rootObject = 64E183582588FD3C00D666BF /* Project object */;
}
`

	appWithTest = `
// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 50;
	objects = {

/* Begin PBXNativeTarget section */
		64E1835F2588FD3C00D666BF /* test */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E1838A2588FD3D00D666BF /* Build configuration list for PBXNativeTarget "test" */;
			buildPhases = (
				64E1835C2588FD3C00D666BF /* Sources */,
				64E1835D2588FD3C00D666BF /* Frameworks */,
				64E1835E2588FD3C00D666BF /* Resources */,
				64E183CA2588FD5F00D666BF /* Embed App Clips */,
				64E183F22588FD9E00D666BF /* Embed App Extensions */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = test;
			productName = test;
			productReference = 64E183602588FD3C00D666BF /* test.app */;
			productType = "com.apple.product-type.application";
		};
		64E183752588FD3D00D666BF /* testTests */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E1838D2588FD3D00D666BF /* Build configuration list for PBXNativeTarget "testTests" */;
			buildPhases = (
				64E183722588FD3D00D666BF /* Sources */,
				64E183732588FD3D00D666BF /* Frameworks */,
				64E183742588FD3D00D666BF /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
				64E183782588FD3D00D666BF /* PBXTargetDependency */,
			);
			name = testTests;
			productName = testTests;
			productReference = 64E183762588FD3D00D666BF /* testTests.xctest */;
			productType = "com.apple.product-type.bundle.unit-test";
		};
/* End PBXNativeTarget section */

/* Begin PBXTargetDependency section */
		64E183782588FD3D00D666BF /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 64E1835F2588FD3C00D666BF /* test */;
			targetProxy = 64E183772588FD3D00D666BF /* PBXContainerItemProxy */;
		};
/* End PBXTargetDependency section */
	};
	rootObject = 64E183582588FD3C00D666BF /* Project object */;
}
`

	appWithTestAndAppClipAndWidget = `
	// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 50;
	objects = {

/* Begin PBXNativeTarget section */
		64E1835F2588FD3C00D666BF /* test */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E1838A2588FD3D00D666BF /* Build configuration list for PBXNativeTarget "test" */;
			buildPhases = (
				64E1835C2588FD3C00D666BF /* Sources */,
				64E1835D2588FD3C00D666BF /* Frameworks */,
				64E1835E2588FD3C00D666BF /* Resources */,
				64E183CA2588FD5F00D666BF /* Embed App Clips */,
				64E183F22588FD9E00D666BF /* Embed App Extensions */,
			);
			buildRules = (
			);
			dependencies = (
				64E183C52588FD5F00D666BF /* PBXTargetDependency */,
				64E183ED2588FD9E00D666BF /* PBXTargetDependency */,
			);
			name = test;
			productName = test;
			productReference = 64E183602588FD3C00D666BF /* test.app */;
			productType = "com.apple.product-type.application";
		};
		64E183752588FD3D00D666BF /* testTests */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E1838D2588FD3D00D666BF /* Build configuration list for PBXNativeTarget "testTests" */;
			buildPhases = (
				64E183722588FD3D00D666BF /* Sources */,
				64E183732588FD3D00D666BF /* Frameworks */,
				64E183742588FD3D00D666BF /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
				64E183782588FD3D00D666BF /* PBXTargetDependency */,
			);
			name = testTests;
			productName = testTests;
			productReference = 64E183762588FD3D00D666BF /* testTests.xctest */;
			productType = "com.apple.product-type.bundle.unit-test";
		};
		64E1839B2588FD5E00D666BF /* clip */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E183C72588FD5F00D666BF /* Build configuration list for PBXNativeTarget "clip" */;
			buildPhases = (
				64E183982588FD5E00D666BF /* Sources */,
				64E183992588FD5E00D666BF /* Frameworks */,
				64E1839A2588FD5E00D666BF /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = clip;
			productName = clip;
			productReference = 64E1839C2588FD5E00D666BF /* clip.app */;
			productType = "com.apple.product-type.application.on-demand-install-capable";
		};
		64E183DC2588FD9D00D666BF /* widgetExtension */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 64E183EF2588FD9E00D666BF /* Build configuration list for PBXNativeTarget "widgetExtension" */;
			buildPhases = (
				64E183D92588FD9D00D666BF /* Sources */,
				64E183DA2588FD9D00D666BF /* Frameworks */,
				64E183DB2588FD9D00D666BF /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = widgetExtension;
			productName = widgetExtension;
			productReference = 64E183DD2588FD9D00D666BF /* widgetExtension.appex */;
			productType = "com.apple.product-type.app-extension";
		};
/* End PBXNativeTarget section */

/* Begin PBXTargetDependency section */
		64E183782588FD3D00D666BF /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 64E1835F2588FD3C00D666BF /* test */;
			targetProxy = 64E183772588FD3D00D666BF /* PBXContainerItemProxy */;
		};
		64E183C52588FD5F00D666BF /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 64E1839B2588FD5E00D666BF /* clip */;
			targetProxy = 64E183C42588FD5F00D666BF /* PBXContainerItemProxy */;
		};
		64E183ED2588FD9E00D666BF /* PBXTargetDependency */ = {
			isa = PBXTargetDependency;
			target = 64E183DC2588FD9D00D666BF /* widgetExtension */;
			targetProxy = 64E183EC2588FD9E00D666BF /* PBXContainerItemProxy */;
		};
/* End PBXTargetDependency section */
	};
	rootObject = 64E183582588FD3C00D666BF /* Project object */;
}
`
)
