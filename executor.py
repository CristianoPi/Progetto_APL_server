import sys
import json
import time
import subprocess
import os
import builtins

# Wrapper per la funzione open per reindirizzare i file di output
class OpenWrapper:
    def __init__(self, base_name):
        self.base_name = base_name
        self.original_open = builtins.open

    def __call__(self, file, mode='r', buffering=-1, encoding=None, errors=None, newline=None, closefd=True, opener=None):
        if 'w' in mode or 'a' in mode:
            file = os.path.join("uploads", f"{self.base_name}_{os.path.basename(file)}")
        return self.original_open(file, mode, buffering, encoding, errors, newline, closefd, opener)

def execute_code(file_path):
    start_time = time.time()
    base_name = os.path.splitext(os.path.basename(file_path))[0]
    output_file_path = os.path.join("uploads", f"{base_name}_output.txt")
    error_file_path = os.path.join("uploads", f"{base_name}_error.txt")

    # Sostituisci la funzione open con il wrapper
    builtins.open = OpenWrapper(base_name)

    try:
        with open(output_file_path, 'w') as output_file, open(error_file_path, 'w') as error_file:
            result = subprocess.run(['python', file_path], stdout=output_file, stderr=error_file, text=True, check=True)
    except subprocess.CalledProcessError as e:
        with open(output_file_path, 'a') as output_file, open(error_file_path, 'a') as error_file:
            output_file.write(e.stdout)
            error_file.write(e.stderr)
    except Exception as e:
        end_time = time.time()
        execution_time = end_time - start_time
        statistics = {
            'execution_time': execution_time,
            'output': '',
            'error': str(e)
        }
        return statistics
    finally:
        # Ripristina la funzione open originale
        builtins.open = open

    end_time = time.time()
    execution_time = end_time - start_time

    with open(output_file_path, 'r') as output_file:
        output = output_file.read()
    with open(error_file_path, 'r') as error_file:
        error = error_file.read()

    statistics = {
        'execution_time': execution_time,
        'output': output,
        'error': error
    }

    return statistics

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({'error': 'File path is required'}))
        sys.exit(1)

    file_path = sys.argv[1]

    # Check if the file exists
    if not os.path.isfile(file_path):
        print(json.dumps({'error': 'File not found'}))
        sys.exit(1)

    # Check if the file is a Python file
    if not file_path.endswith('.py'):
        print(json.dumps({'error': 'Only Python files are allowed'}))
        sys.exit(1)

    stats = execute_code(file_path)
    print(json.dumps(stats))