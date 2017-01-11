#!groovy

import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException
// http://stackoverflow.com/questions/37425064/how-to-use-environment-variables-in-a-groovy-function-using-a-jenkinsfile
import groovy.transform.Field
@Field final BUILD_OS_TARGETS=['el7']

@Field buildParams = [
  "BUILD_OS": "all",
  "REBUILD": "false",
  "LEAVE_CONTAINER": "0",
]
def ask_build_parameter = { ->
  return input(message: "Build Parameters", id: "build_params",
    parameters:[
      [$class: 'ChoiceParameterDefinition',
        choices: "all\n" + BUILD_OS_TARGETS.join("\n"), description: 'Target OS name', name: 'BUILD_OS'],
      [$class: 'ChoiceParameterDefinition',
        choices: "false\ntrue", description: 'Rebuild cache image', name: 'REBUILD'],
      [$class: 'ChoiceParameterDefinition',
        choices: "0\n1", description: 'Leave container after build for debugging.', name: 'LEAVE_CONTAINER'],
    ])
}

// Environment variables supplied by Jenkins system configuration:
// env.REPO_BASE_DIR
// env.BUILD_CACHE_DIR
def write_build_env(label) {
  def build_env="""# These parameters are read from bash and docker --env-file.
# So do not use single or double quote for the value part.
LEAVE_CONTAINER=${buildParams.LEAVE_CONTAINER}
REPO_BASE_DIR=${env.REPO_BASE_DIR ?: ''}
BUILD_CACHE_DIR=${env.BUILD_CACHE_DIR ?: ''}
BUILD_OS=${label}
REBUILD=${buildParams.REBUILD}
RELEASE_SUFFIX=${RELEASE_SUFFIX}
# https://issues.jenkins-ci.org/browse/JENKINS-30252
GIT_BRANCH=${env.BRANCH_NAME}
BRANCH_NAME=${env.BRANCH_NAME}
BRANCH=${env.BRANCH_NAME}
"""
  writeFile(file: "build.env", text: build_env)
}

def checkout_and_merge() {
    checkout scm
    sh "git merge origin/master"
}

@Field RELEASE_SUFFIX=null

def stage_unit_test(label) {
  node(label) {
    stage "Units Tests ${label}"
    checkout_and_merge()
    write_build_env(label)
    sh "./deployment/docker/unit-tests.sh ./build.env"
  }
}

def stage_rpmbuild(label) {
  node(label) {
    stage "RPM Build ${label}"
    checkout_and_merge()
    write_build_env(label)
    sh "./deployment/docker/rpmbuild.sh ./build.env"
  }
}

def stage_integration(label) {
  node("multibox") {
    checkout_and_merge()
    stage "Build Integration Environment"
    write_build_env(label)

    sh "cd ci/multibox/ ; ./build.sh"
    stage "Run Tntegration Test"
    // This is where the integration test will be run
    stage "Cleanup Environment"
    sh "cd ci/multibox/ ; ./destroy.sh --kill"
  }
}

def stage_clean_rpms(label) {
  node(label) {
    checkout_and_merge()
    stage "RPM Repository Cleanups"
    sh "cd repo_cleanups ; ./remove_stale_rpm.sh"
//  sh "cd repo_cleanups ; ./run_remove_stale_rpm.sh"
  }
}

node() {
    stage "Checkout"
    try {
      timeout(time: 10, unit :"SECONDS") {
        buildParams = ask_build_parameter()
      }
    }catch(org.jenkinsci.plugins.workflow.steps.FlowInterruptedException err) {
      // Only ignore errors for timeout.
    }
    checkout_and_merge()
    // http://stackoverflow.com/questions/36507410/is-it-possible-to-capture-the-stdout-from-the-sh-dsl-command-in-the-pipeline
    // https://issues.jenkins-ci.org/browse/JENKINS-26133
    RELEASE_SUFFIX=sh(returnStdout: true, script: "./deployment/packagebuild/gen-dev-build-tag.sh").trim()
}


build_nodes=BUILD_OS_TARGETS.clone()
if( buildParams.BUILD_OS != "all" ){
  build_nodes=[BUILD_OS]
}

// Using .each{} hits "a CPS-transformed closure is not yet supported (JENKINS-26481)"
for( label in build_nodes) {
//stage_unit_test(label)
//stage_rpmbuild(label)
//stage_integration(label)
  stage_clean_rpms(label)
}
