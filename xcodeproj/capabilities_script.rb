require 'xcodeproj'
require 'json'

def read_target(project, target_uuid)
  project.targets.each do |target|
    return target if target.uuid == target_uuid
  end
  nil
end

def read_target_capabilities(project_path)
  project = Xcodeproj::Project.open(project_path)
  raise 'failed to open project' unless project

  root_object = project.root_object
  raise 'project has no root object' unless root_object

  attributes = root_object.attributes
  raise 'project root object has no attributes' unless attributes

  target_attributes = attributes['TargetAttributes']
  raise 'project root object has no target attributes' if target_attributes.nil? || target_attributes.empty?

  target_capabilities = {}
  target_attributes.each do |current_uuid, current_attributes|
    capabilities = current_attributes['SystemCapabilities']
    next if capabilities.nil? || capabilities.empty?

    capability_names = []
    capabilities.each do |key, value|
      enabled = value['enabled']
      next unless enabled || !enabled

      capability_names << key
    end

    target = read_target_name(project, current_uuid)

    target_capabilities[target.name] = capability_names
  end

  target_capabilities
end

begin
  project_path = ENV['project']

  raise 'missing project_path' if project_path.to_s.empty?

  target_capabilities = read_target_capabilities(project_path)

  result = {
    data: target_capabilities
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
