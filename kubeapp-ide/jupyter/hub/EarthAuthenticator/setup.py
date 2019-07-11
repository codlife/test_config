from setuptools import setup

import time

timestamp = time.strftime('%Y%m%d%H%M%S', time.localtime(time.time()))

setup(
    name='earthauthenticator',
    version='0.2.13.' + timestamp,
    author='lucio35',
    author_email="yangzhao16@otcaix.iscas.ac.cn",
    description='CasEarth JupyterHub Authenticator',
    packages=['']
)
