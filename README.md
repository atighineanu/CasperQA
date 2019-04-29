# CasperQA - is an automation tool for an given kubernetes-Caasp cluster 


### options:
 * "new" - <new k8s cluster> - executes all the update commands after the bootstrap (registers, updates, disables the upd.timer and refreshes salt grains); syntax: go run main.go new
 * "supd" - <salt update> - 'transactional-update cleanup dup salt'; syntax: go run main.go supd
 * "pupd" - <package update> 'transactional-update reboot pkg install -y <package name>' (it sets automatically the "ZYPPER_AUTO_IMPORT_KEYS=1" option in /etc/transactional-update.conf; syntax: go run main.go pupd <package name>
 * "cmd" - <command> - salt-cluster <command>;    syntax: go run main.go cmd <command with more arguments>
 * "ar" - <addrepo> - salt-cluster 'zypper ar <your rempo>;  syntax: go run main.go ar <.repo>
