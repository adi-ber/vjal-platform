#!/bin/bash

# === Configuration ===
OUTPUT_FILE="single.txt"
# Set max file size in Megabytes to prevent huge files from being included
MAX_FILE_SIZE_MB=10 # e.g., 10MB

# Directories to completely exclude (paths relative to script location)
# These directories AND their contents will be skipped entirely.
EXCLUDE_DIRS=(
    "./.git"
    "./node_modules"
    "./vendor"
    "./build"
    "./dist"
    "./target"
    "./__pycache__"
    "./*.egg-info"
    "./output" # Excluded .db files found here before
    # --- ADD YOUR PROJECT'S DIRS BELOW ---
    # "./venv"
    # "./data"
    # "./backups"
    # --- End Project Dirs ---
)

# Specific file patterns/names/extensions to exclude CONTENT for.
# These files WILL APPEAR in the directory structure list,
# but their CONTENT will be skipped below. Use BASENAMES or simple patterns.
EXCLUDE_PATTERNS=(
    "*.log"
    "*.tmp"
    "*.swp"
    "*.swo"
    "*.pyc"
    "*.class"
    "*.o"
    "*.a"
    "*.so"
    "*.dll"
    "*.exe"
    "*.png"
    "*.jpg"
    "*.jpeg"
    "*.gif"
    "*.bmp"
    "*.tiff"
    "*.ico"
    "*.pdf"
    "*.zip"
    "*.tar.gz"
    "*.tgz"
    "*.rar"
    "*.jar"
    "*.war"
    # --- ADJUST PATTERNS BELOW BASED ON REQUIREMENTS ---
    "*.DS_Store"      # Exclude content of macOS metadata files
    # "*.enc"         # Keep .enc content
    # "*.json"        # Keep .json content
    "*.db"            # Exclude content of database files
    "*.bak"           # Exclude content of backup files
    "*.ttf"           # Exclude content of font files
    "generate_project_summary.sh" # Exclude content of this script itself
    "single.txt"      # Exclude content of the output file itself
    # --- End Adjustments ---
)
# === End Configuration ===

# --- Argument Parsing ---
DRY_RUN=false
if [[ "$1" == "--dry-run" ]]; then
    DRY_RUN=true
    echo "--- DRY RUN MODE ACTIVATED ---"
    echo "Will list configured exclusions and files that would be processed/skipped for content."
    echo "No output file will be generated."
    echo "Run without '--dry-run' to generate $OUTPUT_FILE."
    echo # Blank line
fi

# --- Verbose Output of Configuration ---
echo "Verbosity: ON"
echo "Output File: $OUTPUT_FILE"
echo "Max File Size Limit: ${MAX_FILE_SIZE_MB}MB"
echo "Excluded Directories (Entirely Skipped):"
printf "  %s\n" "${EXCLUDE_DIRS[@]}"
echo "Excluded File Patterns (Content Skipped):"
printf "  %s\n" "${EXCLUDE_PATTERNS[@]}"
echo # Blank line

# --- Calculate Max Size in Bytes ---
MAX_SIZE_BYTES=$(( MAX_FILE_SIZE_MB * 1024 * 1024 ))

# --- Build `find` arguments ---
# Args for initial structure list (only prune specified dirs)
find_structure_args=(".")
for exclude_dir in "${EXCLUDE_DIRS[@]}"; do
    find_structure_args+=(-path "$exclude_dir" -prune -o)
done
find_structure_args+=(-print) # Print everything not pruned

# Args for getting the list of files for content processing (prune dirs, then select files)
find_files_args=(".") # Start in current dir
for exclude_dir in "${EXCLUDE_DIRS[@]}"; do
    find_files_args+=(-path "$exclude_dir" -prune -o)
done
find_files_args+=(-type f -print0) # Find all files after pruning dirs, print null-separated


# --- Dry Run Execution ---
if [[ "$DRY_RUN" == true ]]; then
    echo "--- Dry Run: Analyzing files for content processing ---"
    echo "(Directory structure is listed separately in full run)"
    echo "(Excluding content for configured patterns and files > ${MAX_FILE_SIZE_MB}MB)"
    echo "*** Binary file detection has been REMOVED ***"
    echo # Blank line

    file_count=0
    skipped_size_count=0
    skipped_pattern_count=0
    processed_count=0

    # Use process substitution '< <(find...)' instead of pipe '|'
    while IFS= read -r -d $'\0' file; do
        ((file_count++))
        filename=$(basename "$file") # *** Get just the filename ***

        # Check Size Limit
        file_size=0
        if stat -c%s "$file" > /dev/null 2>&1; then file_size=$(stat -c%s "$file");
        elif stat -f %z "$file" > /dev/null 2>&1; then file_size=$(stat -f %z "$file"); fi

        if [[ "$file_size" -gt "$MAX_SIZE_BYTES" ]]; then
            echo "[SKIP SIZE]   $file (${file_size} bytes)"
            ((skipped_size_count++))
            continue
        fi

        # Check Excluded Patterns (using filename)
        exclude_this_file=false
        matched_pattern=""
        for pattern in "${EXCLUDE_PATTERNS[@]}"; do
            # *** Compare filename against pattern ***
            if [[ "$filename" == $pattern ]]; then
                exclude_this_file=true
                matched_pattern=$pattern
                break
            fi
        done
        if [[ "$exclude_this_file" == true ]]; then
            echo "[SKIP PATTERN] $file (Filename '$filename' Matches: $matched_pattern)"
            ((skipped_pattern_count++))
            continue
        fi

        # If not skipped by size or pattern, it WILL be processed
        echo "[PROCESS CONTENT] $file"
        ((processed_count++))

    done < <(find "${find_files_args[@]}") # Note '< <(find...)' here

    echo # Blank line
    echo "--- Dry Run Summary ---"
    echo "Total files found for content processing: $file_count"
    echo "Content skipped due to size (> ${MAX_FILE_SIZE_MB}MB): $skipped_size_count"
    echo "Content skipped due to exclude pattern: $skipped_pattern_count"
    echo "Files whose content WOULD be processed: $processed_count"
    echo "--- End Dry Run ---"
    exit 0 # Exit after dry run
fi

# --- Main Execution (Not Dry Run) ---

# Start fresh
echo "Initializing output file: $OUTPUT_FILE"
>"$OUTPUT_FILE"

# Add header info
echo "Project Scan Summary" >> "$OUTPUT_FILE"
echo "Generated: $(date)" >> "$OUTPUT_FILE"
echo "Max File Size Limit Applied: ${MAX_FILE_SIZE_MB}MB" >> "$OUTPUT_FILE"
echo "*** Automatic binary file content detection is DISABLED ***" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Excluded Directories (Entirely Skipped):" >> "$OUTPUT_FILE"
printf "  %s\n" "${EXCLUDE_DIRS[@]}" >> "$OUTPUT_FILE"
echo "Excluded File Patterns (Content Skipped):" >> "$OUTPUT_FILE"
printf "  %s\n" "${EXCLUDE_PATTERNS[@]}" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"


# 1. Add Comprehensive Directory Structure
echo "Generating comprehensive directory structure..."
echo "--- Project Directory Structure ---" >> "$OUTPUT_FILE"
echo "(Lists all files/dirs not inside explicitly excluded directories)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
find "${find_structure_args[@]}" | sort >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "--- End Directory Structure ---" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Directory structure added."


# 2. Add File Contents (with skipping based on size/pattern only)
echo "Generating file contents (applying filters)..."
echo "--- File Contents ---" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

file_count=0
skipped_size_count=0
skipped_pattern_count=0
processed_count=0

# Use process substitution '< <(find...)' instead of pipe '|'
while IFS= read -r -d $'\0' file; do
    ((file_count++))
    filename=$(basename "$file") # *** Get just the filename ***
    echo "Processing: $file"

    # Check Size Limit
    file_size=0
    if stat -c%s "$file" > /dev/null 2>&1; then file_size=$(stat -c%s "$file");
    elif stat -f %z "$file" > /dev/null 2>&1; then file_size=$(stat -f %z "$file");
    else echo "[Warning: Cannot determine size of $file]"; fi

    if [[ "$file_size" -gt "$MAX_SIZE_BYTES" ]]; then
        echo "  Skipping: Size (${file_size} bytes) > Limit (${MAX_FILE_SIZE_MB}MB)"
        echo "--- File: $file ---" >> "$OUTPUT_FILE"
        echo "[Content SKIPPED (size ${file_size} bytes > ${MAX_FILE_SIZE_MB}MB limit)]" >> "$OUTPUT_FILE"
        echo "--- End File: $file ---" >> "$OUTPUT_FILE"; echo "" >> "$OUTPUT_FILE"
        ((skipped_size_count++))
        continue
    fi

    # Check Excluded Patterns (using filename)
    exclude_this_file=false
    matched_pattern=""
    for pattern in "${EXCLUDE_PATTERNS[@]}"; do
         # *** Compare filename against pattern ***
         if [[ "$filename" == $pattern ]]; then
            exclude_this_file=true
            matched_pattern=$pattern
            break
        fi
    done
    if [[ "$exclude_this_file" == true ]]; then
        echo "  Skipping: Filename '$filename' Matches exclude pattern '$matched_pattern'"
        echo "--- File: $file ---" >> "$OUTPUT_FILE"
        echo "[Content SKIPPED (filename '$filename' matches exclude pattern: $matched_pattern)]" >> "$OUTPUT_FILE"
        echo "--- End File: $file ---" >> "$OUTPUT_FILE"; echo "" >> "$OUTPUT_FILE"
         ((skipped_pattern_count++))
        continue
    fi

    # If passed size and pattern checks, append content.

    # Add header to output file
    echo "--- File: $file ---" >> "$OUTPUT_FILE"

    echo "  Appending content..."
    # Use cat -v to handle potential non-printable characters slightly better than raw cat
    cat -v "$file" >> "$OUTPUT_FILE"
    # Ensure newline at the end of the cat output
    if [[ $(tail -c1 "$file" 2>/dev/null | wc -l) -eq 0 ]]; then
       echo "" >> "$OUTPUT_FILE" # Add newline if missing
    fi
    ((processed_count++))


    # Add footer and spacing to output file
    echo "" >> "$OUTPUT_FILE"
    echo "--- End File: $file ---" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"

done < <(find "${find_files_args[@]}") # Note '< <(find...)' here

# Move summary appending to *after* the loop finishes
echo "" >> "$OUTPUT_FILE"
echo "--- Processing Summary ---" >> "$OUTPUT_FILE"
echo "Total files found for content processing step: $file_count" >> "$OUTPUT_FILE"
echo "Content skipped due to size (> ${MAX_FILE_SIZE_MB}MB): $skipped_size_count" >> "$OUTPUT_FILE"
echo "Content skipped due to exclude pattern: $skipped_pattern_count" >> "$OUTPUT_FILE"
echo "Files whose content was processed: $processed_count" >> "$OUTPUT_FILE"
echo "--- End of Summary ---" >> "$OUTPUT_FILE"

# Print summary to console as well (now uses correct counts)
echo # Blank line
echo "--- Processing Summary ---"
echo "Output written to: $OUTPUT_FILE"
echo "Total files found for content processing step: $file_count"
echo "Content skipped due to size (> ${MAX_FILE_SIZE_MB}MB): $skipped_size_count"
echo "Content skipped due to exclude pattern: $skipped_pattern_count"
echo "Files whose content was processed: $processed_count"
echo "--- Done ---"