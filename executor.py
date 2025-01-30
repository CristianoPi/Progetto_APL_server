import os
import json
import sys
import time
import subprocess

def execute_code(file_path):
    start_time = time.time()
    base_name = os.path.splitext(os.path.basename(file_path))[0]
    current_dir = os.path.dirname(os.path.abspath(file_path))

    output_file_path = os.path.join(current_dir, f"{base_name}_output.txt")
    error_file_path = os.path.join(current_dir, f"{base_name}_error.txt")

    # Ottieni i file esistenti e i loro timestamp prima dell'esecuzione
    existing_files = {f: os.path.getmtime(os.path.join(current_dir, f)) for f in os.listdir(current_dir)}

    try:
        # Esegui il subprocess
        result = subprocess.run(["python", file_path], capture_output=True, text=True)
        execution_time = time.time() - start_time

        # Controlla i file creati dopo l'esecuzione
        new_files = []
        for f in os.listdir(current_dir):
            file_path = os.path.join(current_dir, f)
            if f not in existing_files or os.path.getmtime(file_path) > existing_files[f]:
                new_files.append(f)

        result_data = {
            "execution_time": execution_time,
            "errors": result.stderr if result.returncode != 0 else None,
            "created_files": new_files
        }
    except Exception as e:
        execution_time = time.time() - start_time
        result_data = {
            "execution_time": execution_time,
            "errors": str(e),
            "created_files": []
        }

    # Scrivi il risultato nel file di output
    with open(output_file_path, 'w') as output_file:
        json.dump(result_data, output_file)

    return result_data

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