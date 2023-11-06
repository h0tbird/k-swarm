#------------------------------------------------------------------------------
# Make sure we are in the right context
#------------------------------------------------------------------------------

local('kubectl config current-context | grep -q kind-dev')

#------------------------------------------------------------------------------
# Execute another Tiltfile and import named variables into the current scope
#------------------------------------------------------------------------------

load('ext://restart_process', 'docker_build_with_restart')

#------------------------------------------------------------------------------
# Take care of command-line arguments
#------------------------------------------------------------------------------

# tilt up -- --debug
config.define_bool('debug', args=False)

# tilt up -- --flags '--foo=true --bar=true'
config.define_string('flags', args=False)

# Parse the config
cfg = config.parse()
debug = cfg.get('debug', False)
flags = cfg.get('flags', '').split(' ')

#------------------------------------------------------------------------------
# Setup some variables
#------------------------------------------------------------------------------

ARCH = str(local("go env GOARCH")).rstrip("\n")
IMG = 'dev-registry:5000/swarm'

if debug:
  ENTRYPOINT = ['/go/bin/dlv', '--listen=:40000', '--api-version=2', '--headless=true', 'exec', '/manager', '--'] + flags
else:
  ENTRYPOINT = ['/manager'] + flags

DOCKERFILE = '''FROM golang:alpine
RUN apk add gcc musl-dev curl && \
go install github.com/go-delve/delve/cmd/dlv@latest
COPY manager /manager
WORKDIR /
'''

#------------------------------------------------------------------------------
# Build the manager binary
#------------------------------------------------------------------------------

local_resource(
  'binary',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH={ARCH} make build-devel'.format(ARCH=ARCH),
  deps=['internal', 'pkg', 'cmd', 'go.mod', 'go.sum'],
  labels=['manager'],
  allow_parallel=True
)

#------------------------------------------------------------------------------
# Build the docker image
#------------------------------------------------------------------------------

docker_build_with_restart(
  ref=IMG,
  context='./bin',
  entrypoint=ENTRYPOINT,
  live_update=[sync('./bin/manager' , '/manager')],
  dockerfile_contents=DOCKERFILE,
  only=['manager'],
)

#------------------------------------------------------------------------------
# Deploy the yaml
#------------------------------------------------------------------------------

k8s_yaml(local('OVERLAY=dev make overlay'))

#------------------------------------------------------------------------------
# Configure Tilt resources on top automatic assembly
#------------------------------------------------------------------------------

k8s_resource(
  'swarm-controller-manager',
  new_name='deployment',
  labels="manager",
  port_forwards=['40000:40000', '8080:8080'],
)
