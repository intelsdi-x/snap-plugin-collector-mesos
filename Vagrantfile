# This Vagrant environment provisions a single virtual machine that can be
# used to develop and test the Mesos plugin for the Snap telemetry framework.
#
# When provisioning is complete, Mesos will be listening on the following
# addresses and ports:
#
#    * Master: http://10.180.10.180:5050
#    * Agent:  http://10.180.10.180:5051
#
# Set values for the three variables below to install specific versions of
# Mesos, Go, and Snap. Alternately, you may set one or more of these values
# to 'latest' to unpin it.
#
MESOS_RELEASE    = '0.28.2-2.0.27.ubuntu1404'
MARATHON_RELEASE = '1.1.1-1.0.472.ubuntu1404'
GOLANG_RELEASE   = '1.7.3'
SNAP_RELEASE     = '1.0.0'
IP_ADDRESS       = '10.180.10.180'

Vagrant.configure(2) do |config|
  config.vm.box = 'ubuntu/trusty64'
  config.vm.network 'private_network', ip: IP_ADDRESS

  # Forward ports from the guest to the host for ease of access. All services
  # should still be available at IP_ADDRESS:port as well.
  [3000, 5050, 5051, 8080, 8083].each do |port|
    config.vm.network 'forwarded_port', guest: port, host: port
  end

  config.vm.provider 'virtualbox' do |vb|
    vb.name   = 'vagrant-snap-mesos'
    vb.cpus   = 2
    vb.memory = 4096
  end

  config.vm.provision 'shell' do |sh|
    sh.path = 'scripts/provision-vagrant.sh'
    args =  [ '--mesos_release',    MESOS_RELEASE  ]
    args += [ '--marathon_release', MARATHON_RELEASE ]
    args += [ '--golang_release',   GOLANG_RELEASE ]
    args += [ '--snap_release',     SNAP_RELEASE   ]
    args += [ '--ip_address',       IP_ADDRESS     ]
    sh.args = args
  end
end
