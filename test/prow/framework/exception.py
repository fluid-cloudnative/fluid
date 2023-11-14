import os
import sys

project_root = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, project_root)


class TestError(Exception):
    def __init__(self, message):
        self.msg = message
        super(TestError, self).__init__(message)