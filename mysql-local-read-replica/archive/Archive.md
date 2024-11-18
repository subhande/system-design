# Setup MYSQL Read Replica Locally

## Resources
- [MySQL Read Replica Setup Guide](https://channaly.medium.com/setting-up-mysql5-7-read-replica-on-macos-61db2cf600cd)
- [Local MySQL Setup Linux](https://medium.com/@neluwah/setting-up-mysql-replication-for-high-availability-a-step-by-step-guide-fa15e8e5b177)

## Steps

- Config Dir - /opt/homebrew/etc/my.cnf
- Data Dir - /opt/homebrew/var/mysql 

export PATH="/opt/homebrew/bin/mysql/bin:$PATH"

mysqld_safe --defaults-file=/opt/homebrew/etc/my.cnf &


brew install mysql

brew reinstall mysql

brew uninstall mysql

brew services start mysql

brew services restart mysql

brew services stop mysql


sudo lsof -i :3306  


brew services stop mysql
sudo rm -rf /opt/homebrew/bin/mysql
sudo rm -rf /opt/homebrew/etc/my.cnf
sudo rm -rf /opt/homebrew/var/mysql
brew install mysql
brew services start mysql
mysql_secure_installation


 mysql -u root -p 


 mysqld_safe --defaults-file=/opt/homebrew/etc/my.cnf    




SHOW VARIABLES LIKE 'validate_password%';

SHOW VARIABLES LIKE 'validate_password%';
uninstall plugin validate_password;


/opt/homebrew/etc/my-replica.cnf


mysqld --initialize --datadir=/opt/homebrew/var/mysql-replica --user=$(whoami) --explicit_defaults_for_timestamp

mysqld — initialize-insecure --datadir=/opt/homebrew/var/mysql-replica --user=$(whoami) --explicit_defaults_for_timestamp

mysqld_safe --defaults-file=/opt/homebrew/etc/my-replica.cnf &


A temporary password is generated for root@localhost: i#CrHTJJE34W
