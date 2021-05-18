# MySQL Install and Setup for OCTO

## This Ansible role is below run Tasks

- Install MySQL
- Setup start for mysqld
- Create Database OCTO
- Import DDL to DB OCTO

## Need to setup Before Run ansible playbook a dbservers.yml

- Install Ansible(http://docs.ansible.com/ansible/latest/intro_installation.html)
- Create MySQL Server(VM Instance) on GCP(setup network that can access with SSH to mysql server)

## Setup VARS

### Need to set value on group_vars/dbservers

- remote_user : ansible remote user
- server_range : IP addresses ranges of VPC Network of GCP if use the GCP

#### ex)

```~yml
  # dbservers Variables
  for_host:
    # remote user
    remote_user: ansible
  mysqld_5_6:
    # config file for confirm connect to MySQL Server
    defaults_file: /root/.my.cnf
    # OCTO DDL file
    sql_file: /tmp/octo.sql
    # Set IP Range for access to MySQL
    server_range: 10.%
```

### Set DB Server host on inventory/inventory.ini

```~ini
[dbservers]
127.0.0.1
```

## Run the dbservers.yml

```~yml
ansible-playbook -i inventory/inventory.ini dbservers.yml
```
