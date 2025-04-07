package exportoptions

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestManifestIsEmpty(t *testing.T) {
	t.Log("returns true if empty manifest")
	{
		manifest := Manifest{}
		require.Equal(t, true, manifest.IsEmpty())
	}

	t.Log("returns false if not empty manifest")
	{
		manifest := Manifest{
			AppURL: "appURL",
		}
		require.Equal(t, false, manifest.IsEmpty())
	}
	{
		manifest := Manifest{
			DisplayImageURL: "displayImageURL",
		}
		require.Equal(t, false, manifest.IsEmpty())
	}
	{
		manifest := Manifest{
			FullSizeImageURL: "fullSizeImageURL.",
		}
		require.Equal(t, false, manifest.IsEmpty())
	}
	{
		manifest := Manifest{
			AssetPackManifestURL: "assetPackManifestURL.",
		}
		require.Equal(t, false, manifest.IsEmpty())
	}
}

func TestManifestToHash(t *testing.T) {
	t.Log("empty manifest creates empty hash")
	{
		manifest := Manifest{}
		hash := manifest.ToHash()
		require.Equal(t, 0, len(hash))
		{
			value, ok := hash[ManifestAppURLKey]
			require.Equal(t, false, ok)
			require.Equal(t, "", value)
		}
		{
			value, ok := hash[ManifestDisplayImageURLKey]
			require.Equal(t, false, ok)
			require.Equal(t, "", value)
		}
		{
			value, ok := hash[ManifestFullSizeImageURLKey]
			require.Equal(t, false, ok)
			require.Equal(t, "", value)
		}
		{
			value, ok := hash[ManifestAssetPackManifestURLKey]
			require.Equal(t, false, ok)
			require.Equal(t, "", value)
		}
	}

	t.Log("creates hash from manifest")
	{
		manifest := Manifest{
			AppURL:               "appURL",
			DisplayImageURL:      "displayImageURL",
			FullSizeImageURL:     "fullSizeImageURL",
			AssetPackManifestURL: "assetPackManifestURL",
		}
		hash := manifest.ToHash()
		require.Equal(t, 4, len(hash))
		{
			value, ok := hash[ManifestAppURLKey]
			require.Equal(t, true, ok)
			require.Equal(t, "appURL", value)
		}
		{
			value, ok := hash[ManifestDisplayImageURLKey]
			require.Equal(t, true, ok)
			require.Equal(t, "displayImageURL", value)
		}
		{
			value, ok := hash[ManifestFullSizeImageURLKey]
			require.Equal(t, true, ok)
			require.Equal(t, "fullSizeImageURL", value)
		}
		{
			value, ok := hash[ManifestAssetPackManifestURLKey]
			require.Equal(t, true, ok)
			require.Equal(t, "assetPackManifestURL", value)
		}
	}
}

func TestNewAppStoreConnectOptions(t *testing.T) {
	t.Log("create app-store type export options with default values")
	{
		options := NewAppStoreConnectOptions()
		require.Equal(t, UploadBitcodeDefault, options.UploadBitcode)
		require.Equal(t, UploadSymbolsDefault, options.UploadSymbols)
		require.Equal(t, TestFlightInternalTestingOnlyDefault, options.TestFlightInternalTestingOnly)
	}
}

func TestAppStoreOptionsToHash(t *testing.T) {
	t.Log("default app-store type options creates hash with legacy method")
	{
		options := NewAppStoreOptions()
		options.ManageAppVersion = true
		hash := options.Hash()
		require.Equal(t, 1, len(hash), fmt.Sprintf("Hash: %+v", hash))

		{
			value, ok := hash[MethodKey]
			require.Equal(t, true, ok)
			require.Equal(t, MethodAppStore, value)
		}
	}

	t.Log("default app-store type options creates hash with new method")
	{
		options := NewAppStoreConnectOptions()
		options.ManageAppVersion = true
		hash := options.Hash()
		require.Equal(t, 1, len(hash), fmt.Sprintf("Hash: %+v", hash))

		{
			value, ok := hash[MethodKey]
			require.Equal(t, true, ok)
			require.Equal(t, MethodAppStoreConnect, value)
		}
	}

	t.Log("custom app-store type option's generated hash contains all properties")
	{
		options := NewAppStoreOptions()
		options.TeamID = "123"
		options.UploadBitcode = false
		options.UploadSymbols = false
		options.ManageAppVersion = false
		options.TestFlightInternalTestingOnly = true

		hash := options.Hash()
		require.Equal(t, 6, len(hash))

		{
			value, ok := hash[MethodKey]
			require.True(t, ok)
			require.Equal(t, MethodAppStore, value)
		}
		{
			value, ok := hash[TeamIDKey]
			require.True(t, ok)
			require.Equal(t, "123", value)
		}
		{
			value, ok := hash[UploadBitcodeKey]
			require.True(t, ok)
			require.Equal(t, false, value)
		}
		{
			value, ok := hash[UploadSymbolsKey]
			require.True(t, ok)
			require.Equal(t, false, value)
		}
		{
			value, ok := hash[manageAppVersionKey]
			require.True(t, ok)
			require.Equal(t, false, value)
		}
		{
			value, ok := hash[TestFlightInternalTestingOnlyKey]
			require.True(t, ok)
			require.Equal(t, true, value)
		}
	}
}

func TestAppStoreOptionsWriteToFile(t *testing.T) {
	t.Log("default app-store type options overrides only method")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
		require.NoError(t, err)
		pth := filepath.Join(tmpDir, "exportOptions.plist")

		options := NewAppStoreConnectOptions()
		options.ManageAppVersion = true
		require.NoError(t, options.WriteToFile(pth))

		content, err := fileutil.ReadStringFromFile(pth)
		require.NoError(t, err)
		desired := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>method</key>
		<string>app-store-connect</string>
	</dict>
</plist>`
		require.Equal(t, desired, content)
	}

	t.Log("custom app-store type options overrides all properties")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
		require.NoError(t, err)
		pth := filepath.Join(tmpDir, "exportOptions.plist")

		options := NewAppStoreOptions()
		options.TeamID = "123"
		options.UploadBitcode = false
		options.UploadSymbols = false
		options.ManageAppVersion = false
		require.NoError(t, options.WriteToFile(pth))

		content, err := fileutil.ReadStringFromFile(pth)
		require.NoError(t, err)
		desired := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>manageAppVersionAndBuildNumber</key>
		<false/>
		<key>method</key>
		<string>app-store</string>
		<key>teamID</key>
		<string>123</string>
		<key>uploadBitcode</key>
		<false/>
		<key>uploadSymbols</key>
		<false/>
	</dict>
</plist>`
		require.Equal(t, desired, content)
	}
}

func TestNonNewAppStoreOptions(t *testing.T) {
	t.Log("create NON app-store type export options with default values")
	{
		options := NewNonAppStoreOptions(MethodDevelopment, false)
		require.Equal(t, MethodDevelopment, options.Method)
		require.Equal(t, CompileBitcodeDefault, options.CompileBitcode)
		require.Equal(t, EmbedOnDemandResourcesAssetPacksInBundleDefault, options.EmbedOnDemandResourcesAssetPacksInBundle)
		require.Equal(t, ICloudContainerEnvironment(""), options.ICloudContainerEnvironment)
		require.Equal(t, ThinningDefault, options.Thinning)
	}
}

func TestNonAppStoreOptionsToHash(t *testing.T) {
	t.Log("default NON app-store type options creates hash with method")
	{
		options := NewNonAppStoreOptions(MethodDevelopment, false)
		hash := options.Hash()
		require.Equal(t, 1, len(hash))

		{
			value, ok := hash[MethodKey]
			require.Equal(t, true, ok)
			require.Equal(t, MethodDevelopment, value)
		}
	}

	t.Log("custom NON app-store type option's generated hash contains all properties")
	{
		options := NewNonAppStoreOptions(MethodEnterprise, false)
		options.TeamID = "123"
		options.CompileBitcode = false
		options.EmbedOnDemandResourcesAssetPacksInBundle = false
		options.ICloudContainerEnvironment = ICloudContainerEnvironmentProduction
		options.OnDemandResourcesAssetPacksBaseURL = "url"
		options.Thinning = ThinningThinForAllVariants
		options.Manifest = Manifest{
			AppURL:               "appURL",
			DisplayImageURL:      "displayImageURL",
			FullSizeImageURL:     "fullSizeImageURL",
			AssetPackManifestURL: "assetPackManifestURL",
		}

		hash := options.Hash()
		require.Equal(t, 8, len(hash))

		{
			value, ok := hash[MethodKey]
			require.Equal(t, true, ok)
			require.Equal(t, MethodEnterprise, value)
		}
		{
			value, ok := hash[TeamIDKey]
			require.Equal(t, true, ok)
			require.Equal(t, "123", value)
		}
		{
			value, ok := hash[CompileBitcodeKey]
			require.Equal(t, true, ok)
			require.Equal(t, false, value)
		}
		{
			value, ok := hash[EmbedOnDemandResourcesAssetPacksInBundleKey]
			require.Equal(t, true, ok)
			require.Equal(t, false, value)
		}
		{
			value, ok := hash[ICloudContainerEnvironmentKey]
			require.Equal(t, true, ok)
			require.Equal(t, ICloudContainerEnvironmentProduction, value)
		}
		{
			value, ok := hash[OnDemandResourcesAssetPacksBaseURLKey]
			require.Equal(t, true, ok)
			require.Equal(t, "url", value)
		}
		{
			value, ok := hash[ThinningKey]
			require.Equal(t, true, ok)
			require.Equal(t, ThinningThinForAllVariants, value)
		}
		{
			manifestHash, ok := hash[ManifestKey].(map[string]string)
			require.Equal(t, true, ok)
			require.Equal(t, 4, len(manifestHash))

			{
				value, ok := manifestHash[ManifestAppURLKey]
				require.Equal(t, true, ok)
				require.Equal(t, "appURL", value)
			}
			{
				value, ok := manifestHash[ManifestDisplayImageURLKey]
				require.Equal(t, true, ok)
				require.Equal(t, "displayImageURL", value)
			}
			{
				value, ok := manifestHash[ManifestFullSizeImageURLKey]
				require.Equal(t, true, ok)
				require.Equal(t, "fullSizeImageURL", value)
			}
			{
				value, ok := manifestHash[ManifestAssetPackManifestURLKey]
				require.Equal(t, true, ok)
				require.Equal(t, "assetPackManifestURL", value)
			}
		}
	}
}

func TestNonAppStoreOptionsWriteToFile(t *testing.T) {
	t.Log("default NON app-store type options overrides only method")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
		require.NoError(t, err)
		pth := filepath.Join(tmpDir, "exportOptions.plist")

		options := NewNonAppStoreOptions(MethodEnterprise, false)
		require.NoError(t, options.WriteToFile(pth))

		content, err := fileutil.ReadStringFromFile(pth)
		require.NoError(t, err)
		desired := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>method</key>
		<string>enterprise</string>
	</dict>
</plist>`
		require.Equal(t, desired, content)
	}

	t.Log("custom app-store type options overrides all properties")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
		require.NoError(t, err)
		pth := filepath.Join(tmpDir, "exportOptions.plist")

		options := NewNonAppStoreOptions(MethodEnterprise, false)
		options.TeamID = "123"
		options.CompileBitcode = false
		options.EmbedOnDemandResourcesAssetPacksInBundle = false
		options.ICloudContainerEnvironment = ICloudContainerEnvironmentProduction
		options.OnDemandResourcesAssetPacksBaseURL = "url"
		options.Thinning = ThinningThinForAllVariants
		options.Manifest = Manifest{
			AppURL:               "appURL",
			DisplayImageURL:      "displayImageURL",
			FullSizeImageURL:     "fullSizeImageURL",
			AssetPackManifestURL: "assetPackManifestURL",
		}

		require.NoError(t, options.WriteToFile(pth))

		content, err := fileutil.ReadStringFromFile(pth)
		require.NoError(t, err)
		desired := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>compileBitcode</key>
		<false/>
		<key>embedOnDemandResourcesAssetPacksInBundle</key>
		<false/>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>manifest</key>
		<dict>
			<key>appURL</key>
			<string>appURL</string>
			<key>assetPackManifestURL</key>
			<string>assetPackManifestURL</string>
			<key>displayImageURL</key>
			<string>displayImageURL</string>
			<key>fullSizeImageURL</key>
			<string>fullSizeImageURL</string>
		</dict>
		<key>method</key>
		<string>enterprise</string>
		<key>onDemandResourcesAssetPacksBaseURL</key>
		<string>url</string>
		<key>teamID</key>
		<string>123</string>
		<key>thinning</key>
		<string>thin-for-all-variants</string>
	</dict>
</plist>`
		require.Equal(t, desired, content)
	}
}
