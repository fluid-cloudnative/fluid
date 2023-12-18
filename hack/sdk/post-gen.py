import os
import argparse

def handle_kubernetes_import(python_sdk_path):
    kubernetes_imports = """# Import Kubernetes models.
from kubernetes.client import *

"""
    model_module_init_py_file = os.path.join(python_sdk_path, "fluid", "models", "__init__.py")

    lines = []
    with open(model_module_init_py_file, "r") as f:
        lines = f.readlines()
        print(lines)

    with open(model_module_init_py_file, "w") as f:
        for line in lines:
            if line == "# import models into model package\n":
                # Insert kubernetes models into fluid sdk pacakge
                f.write(kubernetes_imports)
            f.write(line)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Post-gen script for generating Fluid SDK')
    parser.add_argument("-p", "--python-sdk-path", help="Where does the auto-generated python sdk stores.")
    args = parser.parse_args()

    handle_kubernetes_import(args.python_sdk_path)
