echo "Configuring vagrant-specific stuff ..."

# Vagrant specific
date > /etc/vagrant_box_build_time

# install wget
yum -y install wget

# disable iptables
/etc/init.d/iptables stop
/sbin/chkconfig iptables off

# Add vagrant user
/usr/sbin/groupadd vagrant
/usr/sbin/useradd vagrant -g vagrant -G wheel
echo "vagrant"|passwd --stdin vagrant
/bin/sed -i 's/[^\!]requiretty/\!requiretty/' /etc/sudoers 
/bin/sed -i 's/^\(Default.*secure_path.*$\)/\1:\/usr\/local\/bin/' /etc/sudoers
echo "vagrant        ALL=(ALL)       NOPASSWD: ALL" >> /etc/sudoers.d/vagrant
chmod 0440 /etc/sudoers.d/vagrant

# Speed up ssh connections
/bin/sed -i 's/^#UseDNS.*$/UseDNS no/' /etc/ssh/sshd_config
/bin/sed -i 's/^.*GSSAPIAuthentication.*$/GSSAPIAuthentication no/' /etc/ssh/sshd_config

# Installing vagrant keys
mkdir -pm 700 /home/vagrant/.ssh
wget --no-check-certificate 'https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub' -O /home/vagrant/.ssh/authorized_keys
chmod 0600 /home/vagrant/.ssh/authorized_keys
chown -R vagrant:vagrant /home/vagrant/.ssh
 