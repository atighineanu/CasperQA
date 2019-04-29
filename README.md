# CasperQA - is an automation tool for an given kubernetes-Caasp cluster 


### options:
  "new" - <new k8s cluster> - executes all the update commands after the bootstrap (registers, updates, disables the upd.timer and refreshes salt grains)
  "supd" - <salt update> - 'transactional-update cleanup dup salt'
  "pupd" - <package update> 'transactional-update reboot pkg install -y <package name>' (it sets automatically the "ZYPPER_AUTO_IMPORT_KEYS=1" option in /etc/transactional-update.conf
