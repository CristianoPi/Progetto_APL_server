import sys
import json
import time
import subprocess
import os
import builtins

# Wrapper per reindirizzare i file di output
class OpenWrapper:
    def __init__(self, base_name):
        self.base_name = base_name
        self.original_open = builtins.open
        self.uploads_dir = os.path.join("uploads")
        os.makedirs(self.uploads_dir, exist_ok=True)  # Crea la directory

    def __call__(self, file, *args, **kwargs):
        if any(mode in kwargs.get('mode', '') for mode in ('w', 'a', 'x')):
            file = os.path.join(self.uploads_dir, f"{self.base_name}_{os.path.basename(file)}")
        return self.original_open(file, *args, **kwargs)

def execute_code(file_path):
    start_time = time.time()
    base_name = os.path.splitext(os.path.basename(file_path))[0]
    
    # Crea la directory uploads (doppio check)
    uploads_dir = os.path.join("uploads")
    os.makedirs(uploads_dir, exist_ok=True)

    output_file_path = os.path.join(uploads_dir, f"{base_name}_output.txt")
    error_file_path = os.path.join(uploads_dir, f"{base_name}_error.txt")

    # Sostituzione della funzione open
    builtins.open = OpenWrapper(base_name)

    try:
        with open(output_file_path, 'w') as output_file, open(error_file_path, 'w') as error_file:
            result = subprocess.run(
                ['python', file_path],
                stdout=output_file,
                stderr=error_file,
                text=True,
                check=True
            )
    except subprocess.CalledProcessError as e:
        pass  # Gli errori sono già scritti nei file
    except Exception as e:
        end_time = time.time()
        return {
            'execution_time': end_time - start_time,
            'output': '',
            'error': str(e)
        }
    finally:
        builtins.open = open  # Ripristino

    end_time = time.time()

    # Lettura sicura dei file
    def safe_read(path):
        try:
            with open(path, 'r') as f:
                return f.read()
        except FileNotFoundError:
            return ""

    output = safe_read(output_file_path)
    error = safe_read(error_file_path)

    return {
        'execution_time': end_time - start_time,
        'output': output,
        'error': error
    }

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({'error': 'File path is required'}))
        sys.exit(1)

    file_path = sys.argv[1]

    if not os.path.isfile(file_path):
        print(json.dumps({'error': 'File not found'}))
        sys.exit(1)

    if not file_path.endswith('.py'):
        print(json.dumps({'error': 'Only Python files are allowed'}))
        sys.exit(1)

    stats = execute_code(file_path)
    print(json.dumps(stats, indent=2))