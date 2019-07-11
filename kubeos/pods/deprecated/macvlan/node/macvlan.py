#####################################################
## Copyright (2019, ) Institute of Software
##        Chinese Academy of Sciences
##         wuheng@otcaix.iscsa.ac.cn
#####################################################

import os
import time
import logging
from shutil import copyfile
import ConfigParser
from datetime import datetime



logging.basicConfig(level=logging.INFO,filename='/var/log/macvlan.log')
config = ConfigParser.ConfigParser()
config.read("cni.conf")
    
# MACVLAN_NIC
def getNIC():  
    nic = os.getenv(config.get("variable", "varNic"), '')
    if nic == '':
        logging.error("%s Cannot find available IP" % datetime.now())
        exit
    return str.strip(nic)      


# MACVLAN_GATEWAY
def getGateway(nic):
    
    gw = os.getenv(config.get("variable", "varGateway"), 
                   os.popen(config.get("command", "findGatewayCmd").replace('NIC', nic)).read())
    
    if gw is None or gw == '':
        logging.error("%s Cannot find Gateway" % datetime.now())
        exit
    elif cmp(gw, 'localhost'):
        gw = '127.0.0.1'
      
    return str.strip(gw)
           
# MACVLAN_SUBNET
def getSubnet(nic):
    
    sn = os.getenv(config.get("variable", "varSubnet"), 
                   os.popen(config.get("command", "findSubnetCmd").replace('NIC', nic)).read())

    if sn is None or sn == '':
        logging.error ("%s Cannot find subnet" % datetime.now())
        exit
        
    return str.strip(sn).replace('/', '\/')


# MACVLAN_RANGESTART
# MACVLAN_RANGEEND
def getIPRanges(subnet):
    
    strs = subnet.split('/')
    ip = strs[0]
    mask = 32 - int(strs[1])
    
    if mask < 8 or mask >= 32:
        logging.error ("%s Invalid mask: %s" % (datetime.now(), mask))
        exit
      
    ipPrefix = ip
    while mask > 0:
        idx = ipPrefix.rfind('.')
        ipPrefix = ipPrefix[0:idx]
        mask = mask - 8;
    
    length = len(ipPrefix.split('.'))
    rangeStart = ipPrefix
    rangeEnd = ipPrefix
    
    while length < 4:
        rangeStart = rangeStart + ".1"
        rangeEnd = rangeEnd + ".240" 
        length = length + 1
      
    return (os.getenv(config.get("variable", "varStartIP"), str.strip(rangeStart))
            , os.getenv(config.get("variable", "varEndIP"), str.strip(rangeEnd)))

def getValues():
    nic = getNIC()
    gw = getGateway(nic)
    sn = getSubnet(nic)
    rs, re = getIPRanges(sn)
    logging.info("%s NIC name: %s, Subnet: %s, Gateway: %s, RangeStart: %s, RangeEnd: %s" 
                 % (datetime.now(), nic, sn, gw, rs, re))
    return nic, sn, gw, rs, re

def enableCNI():
    
    binDir = config.get("config", "binDir")
    if not os.path.exists(binDir):
        os.popen('mkdir -p ' + binDir)
        logging.info("%s mkdir  %s" % (datetime.now(), binDir))
    
    
    for binFile in config.get("config", "binFiles").split(","):
        if not os.path.exists(binDir + "/" + binFile):
            copyfile(binFile, binDir + "/" + binFile)
            os.chmod(binDir + "/" + binFile, 777)
            logging.info("%s enable macvlan bin file:  %s" % (datetime.now(), binDir + "/" + binFile))
      
        
    configDir = config.get("config", "configDir")
    if not os.path.exists(configDir):
        os.popen('mkdir -p ' + configDir)
        logging.info("% mkdir  %s" % (datetime.now(), configDir))
        
    configFile = config.get("config", "configFile")
    if not os.path.exists(configDir + "/" + configFile):
        
        copyfile(configFile + ".template", configDir + "/" + configFile)
        os.chmod(configDir + "/" + configFile, 777);
        
        nic, sn, gw, rs, re = getValues() 
        cmd = config.get("command", "replaceTextCmd")
        os.popen(cmd.replace('OLD', 'NIC')
                 .replace('NEW', nic)
                 .replace('FILE', configDir + "/" + configFile))
        os.popen(cmd.replace('OLD', 'SUBNET')
                 .replace('NEW', sn)
                 .replace('FILE', configDir + "/" + configFile))
        os.popen(cmd.replace('OLD', 'GW')
                 .replace('NEW', gw)
                 .replace('FILE', configDir + "/" + configFile))
        os.popen(cmd.replace('OLD', 'START')
                 .replace('NEW', rs)
                 .replace('FILE', configDir + "/" + configFile))
        os.popen(cmd.replace('OLD', 'END')
                 .replace('NEW', re)
                 .replace('FILE', configDir + "/" + configFile))
        logging.info("%s enable macvlan config at  %s" % (datetime.now(), configDir + "/" + configFile))


           
if __name__ == '__main__':

    while True:
        enableCNI()
        time.sleep(10)

    
    
    
