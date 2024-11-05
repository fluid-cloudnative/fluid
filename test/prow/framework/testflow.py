import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, project_root)

from framework.exception import TestError


class TestFlow():
    def __init__(self, case):
        self.case = case
        self.steps = []

    def append_step(self, step):
        self.steps.append(step)

    def run(self):
        print("> Testcase \"{}\" started <".format(self.case))
        undergoing_step = 0
        total_steps = len(self.steps)
        failed = False
        try:
            while undergoing_step < total_steps:
                self.steps[undergoing_step].go_forth()
                print("PASS {}".format(self.steps[undergoing_step].get_step_name()))
                time.sleep(3)
                undergoing_step += 1
        except TestError as e:
            failed = True
            print("FAIL {}".format(self.steps[undergoing_step].get_step_name()))
            msg = e.msg
            raise Exception("> Testcase \"{}\" failed at Step \"{}\": {}".format(self.case, self.steps[undergoing_step].get_step_name(), msg))
        except Exception as e:
            failed = True
            print("FAIL {}".format(self.steps[undergoing_step].get_step_name()))
            raise e
        finally:
            if undergoing_step >= total_steps:
                undergoing_step = total_steps - 1
            while undergoing_step >= 0:
                self.steps[undergoing_step].go_back()
                undergoing_step -= 1
            if not failed:
                print("> Testcase \"{}\" succeeded <\n\n".format(self.case))

