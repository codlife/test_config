"""Earth Authenticator for JupyterHub
"""

from jupyterhub.auth import Authenticator
import pymysql
from tornado import gen


class EarthAuthenticator(Authenticator):
    config = {
        "database": "bdp",
        "user": "root",
        "password": "Oncedi@2018",
        "host": "39.104.74.209",
        "port": 3306,
        "charset": "utf8"
    }
    conn = pymysql.connect(**config)
    cur = conn.cursor()

    @gen.coroutine
    def authenticate(self, handler, data):
        userId = data['username']
        passwd = data['password']
        if userId == None or passwd == None or len(userId) == 0 or len(passwd) == 0:
            return None
        self.reConnect()
        self.cur = self.conn.cursor()
        self.cur.execute("select password from MainUser where id = '" + userId + "'")
        rows = self.cur.fetchall()
        if len(rows) == 0 or rows[0][0] != passwd:
            return None

        return data['username']

    @gen.coroutine
    def reConnect(self):
        try:
            self.conn.ping()
        except:
            self.conn = pymysql.connect(**self.config)
