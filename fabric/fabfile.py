#!/usr/bin/env python
from fabric.api import env, local, run, sudo
from fabric.operations import put

env.hosts = ['picostats@46.101.254.59']


def deploy():
    local('mkdir picostatsnew')
    local('cp ../picostats picostatsnew/')
    local('cp -R ../public picostatsnew/')
    local('cp -R ../templates picostatsnew/')
    local('cp production.json picostatsnew/config.json')
    local('tar -czf picostatsnew.tar.gz picostatsnew/')
    local('rm -rf picostatsnew')
    put('picostatsnew.tar.gz')
    run('tar -xzf picostatsnew.tar.gz')
    run('rm -rf picostats')
    run('mv picostatsnew picostats')
    run('rm picostatsnew.tar.gz')
    local('rm picostatsnew.tar.gz')
    sudo('supervisorctl reload')
