import os
import argparse

REPO_ROOT = os.path.realpath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".."))

# The only Python SDK that hack/sdk/gen-sdk.sh generates (PYTHON_SDK_OUTPUT_PATH).
PYTHON_SDK_DIRS = (os.path.join(REPO_ROOT, "sdk", "python"),)


def resolve_sdk_path(python_sdk_path):
    # Only rewrite a known SDK directory, and use the known path rather than the argument from here on.
    requested = os.path.realpath(python_sdk_path)
    for sdk_dir in PYTHON_SDK_DIRS:
        if requested == os.path.realpath(sdk_dir):
            return sdk_dir
    raise SystemExit("--python-sdk-path must be one of %s, got %s" % (", ".join(PYTHON_SDK_DIRS), python_sdk_path))


def handle_kubernetes_import(python_sdk_path):
    kubernetes_imports = """# Import Kubernetes models.
from kubernetes.client import *

"""
    model_module_init_py_file = os.path.join(python_sdk_path, "fluid", "models", "__init__.py")
    if not os.path.isfile(model_module_init_py_file):
        raise SystemExit("%s not found, is %s the generated python SDK?" % (model_module_init_py_file, python_sdk_path))

    lines = []
    with open(model_module_init_py_file, "r") as f:
        lines = f.readlines()
        print(lines)

    with open(model_module_init_py_file, "w") as f:
        for line in lines:
            if line == "# import models into model package\n":
                # Insert kubernetes models into fluid sdk package
                f.write(kubernetes_imports)
            f.write(line)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Post-gen script for generating Fluid SDK')
    parser.add_argument("-p", "--python-sdk-path", required=True, help="Where does the auto-generated python sdk stores.")
    args = parser.parse_args()

    handle_kubernetes_import(resolve_sdk_path(args.python_sdk_path))
