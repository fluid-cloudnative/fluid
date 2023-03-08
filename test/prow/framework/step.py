import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, project_root)

from framework.exception import TestError


def check(fn, retries, interval):
    def check_internal():
        tt = 0
        while tt < retries:
            if fn():
                return
            time.sleep(interval)
            tt += 1

        raise TestError("timeout for {} seconds".format(retries * interval))

    return check_internal


def dummy_back():
    pass


class SimpleStep():
    def __init__(self, step_name, forth_fn, back_fn):
        self.step_name = step_name
        self.forth_fn = forth_fn
        self.back_fn = back_fn

    def get_step_name(self):
        return self.step_name

    def go_forth(self):
        self.forth_fn()

    def go_back(self):
        self.back_fn()


class StatusCheckStep(SimpleStep):
    def __init__(self, step_name, forth_fn, back_fn=dummy_back, timeout=120, interval=1):
        super().__init__(step_name, check(forth_fn, timeout, interval), back_fn)