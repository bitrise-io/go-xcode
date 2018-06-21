package xcodeproj

import "github.com/bitrise-tools/go-xcode/xcodeproj/ruby"

// SchemeModel ...
type SchemeModel struct {
	Name         string
	IsTestable   bool
	IsArchivable bool
}

// ProjectSharedSchemes ...
func ProjectSharedSchemes(projectPth string) ([]SchemeModel, error) {
	// return sharedSchemes(projectPth)
	runner := ruby.NewRunner(schemesRubyScript, map[string]string{"project": projectPth})
	if err := runner.BundleInstall(map[string]string{"xcodeproj": "", "json": ""}); err != nil {
		return nil, err
	}

	type output struct {
		ArchivableSchemes []string
		TestableSchemes   []string
	}
	var out output
	if err := runner.Execute(&out); err != nil {
		return nil, err
	}

	schemesByName := map[string]SchemeModel{}
	for _, schemeName := range out.ArchivableSchemes {
		scheme, ok := schemesByName[schemeName]
		if !ok {
			scheme = SchemeModel{Name: schemeName}
		}
		scheme.IsArchivable = true
		schemesByName[schemeName] = scheme
	}
	for _, schemeName := range out.TestableSchemes {
		scheme, ok := schemesByName[schemeName]
		if !ok {
			scheme = SchemeModel{Name: schemeName}
		}
		scheme.IsTestable = true
		schemesByName[schemeName] = scheme
	}

	var schemes []SchemeModel
	for _, scheme := range schemesByName {
		schemes = append(schemes, scheme)
	}

	return schemes, nil
}

const schemesRubyScript = `require 'xcodeproj'
require 'json'

def archivable_scheme?(scheme)
  action = scheme.build_action
  return false unless action

  entries = action.entries || []
  return false if entries.empty?

  entries = entries.select(&:build_for_archiving?) || []
  !entries.empty?
end

def testable_scheme?(scheme)
  action = scheme.test_action
  return false unless action

  testables = action.testables || []
  return false if testables.empty?

  testables = testables.select { |ref| !ref.skipped? }
  !testables.empty?
end

def shared_schemes_by_project(project_path)
  archivable_schemes = []
  testable_schemes = []

  Dir.glob(File.join(project_path, 'xcshareddata', 'xcschemes', '*.xcscheme')).each do |scheme_path|
    scheme = Xcodeproj::XCScheme.new(scheme_path)

    if archivable_scheme?(scheme)
      archivable_schemes << File.basename(scheme_path, '.xcscheme')
    end

    if testable_scheme?(scheme)
      testable_schemes << File.basename(scheme_path, '.xcscheme')
    end
  end

  [archivable_schemes, testable_schemes]
end

begin
  project_path = ENV['project']

  raise 'missing project_path' if project_path.to_s.empty?

  archivable_schemes, testable_schemes = shared_schemes_by_project(project_path)
  result = {
    data: {
      archivable_schemes: archivable_schemes,
      testable_schemes: testable_schemes,
    }
  }
  result_json = JSON.pretty_generate(result).to_s
  puts result_json
rescue => e
  error_message = e.to_s + "\n" + e.backtrace.join("\n")
  result = {
    error: error_message
  }
  result_json = result.to_json.to_s
  puts result_json
  exit(1)
end
`
