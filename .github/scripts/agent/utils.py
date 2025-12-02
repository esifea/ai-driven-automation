import os

def parse_files(response_text):
    """Pure function: Parses the AI response into a dictionary."""
    files = {}
    lines = response_text.split('\n')
    current_file = None
    code_lines = []
    in_block = False

    for line in lines:
        if line.strip().startswith("### File:"):
            if current_file and code_lines:
                files[current_file] = "\n".join(code_lines).strip()
            current_file = line.split(":", 1)[1].strip()
            code_lines = []
            in_block = False
        elif line.strip().startswith("```"):
            in_block = not in_block
        elif in_block or (current_file and not line.strip().startswith("```")):
            code_lines.append(line)
    
    if current_file and code_lines:
        files[current_file] = "\n".join(code_lines).strip()
    
    return files

def write_files_to_disk(files_dict):
    """Side-effect function: Writes the dictionary to disk."""
    count = 0
    for path, content in files_dict.items():
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, 'w') as f:
            f.write(content)
        print(f"Wrote: {path}")
        count += 1
    return count
