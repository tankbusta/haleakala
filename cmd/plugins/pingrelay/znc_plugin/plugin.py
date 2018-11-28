import json
import time
import traceback

from datetime import datetime

try:
    import znc
    import zmq
except ImportError:
    raise Exception("You are missing ZMQ or are not running this script within the context of ZNC")


DEFAULT_SOCKET = 'tcp://0.0.0.0:5555'
#DEFAULT_SOCKET = 'ipc:///tmp/irc_messages'

class zmqrelay(znc.Module):
    description = "Relay FleetBot ping's to Discord"
    module_types = (znc.CModInfo.NetworkModule,)

    def _write_log(self, msg):
        now = datetime.now().strftime("%a, %d %b %Y %H:%M:%S")
        with open('/tmp/relay_debug_log.txt', 'ab') as fd:
            fd.write('{:s} - {}\n'.format(now, msg).encode('utf-8'))

    def _load_socket(self):
        try:
            self.ctx = zmq.Context()
            self._write_log(self.ctx)

            self.sock = self.ctx.socket(zmq.PUB)
            self.sock.bind(DEFAULT_SOCKET)
            self._write_log(self.sock)
        except Exception:
            self._write_log(traceback.format_exc())

    def OnLoad(self, *args):
        self._load_socket()
        self._write_log('module has been loaded')
        return True

    def OnModuleUnloading(self, *args):
        #self.sock.close()
        self._write_log('module unloading')
        return True

    def OnPrivMsg(self, nick, zmessage):
        msg = zmessage.s
        from_user = nick.GetNick()
        self._write_log('msg recv {:s} from {:s}'.format(msg, from_user))

        try:
            msg = {'nick': nick.GetNick(), 'message': zmessage.s, 'now': time.time()}
            self.sock.send_string(json.dumps(msg), zmq.NOBLOCK)
            self._write_log('send msg over zmq')
        except Exception:
            self._write_log(traceback.format_exc())

        return znc.CONTINUE
