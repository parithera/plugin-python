#!/bin/bash

. ./config.sh

# Define variables
bam_filename="$outputdir/$name.bam"
star_reference_dir="indexes/$reference/index"
sjdbgtf="indexes/$reference/file.gtf"

r1_search="fastq/$experiment/*R1*"
r2_search="fastq/$experiment/*R2*"
files_with_r1=$(find $r1_search -printf "%p," | sed 's/,$//')
files_with_r2=$(find $r2_search -printf "%p," | sed 's/,$//')

if $inverted ; then
    files_with_r2=$(find $r1_search -printf "%p," | sed 's/,$//')
    files_with_r1=$(find $r2_search -printf "%p," | sed 's/,$//')
fi

echo "Files with R1: $files_with_r1"
echo "Files with R2: $files_with_r2"

whitelists_folder="$whitelists_folder/*"
whitelists=$(find $whitelists_folder -printf "%p ")

echo "Whitelists: $whitelists \n\n"

# Run STAR
STAR \
--runThreadN $threads \
--runMode alignReads \
--outSAMtype BAM Unsorted \
--sysShell /bin/bash \
--genomeDir "$star_reference_dir" \
--readFilesIn "$files_with_r1" "$files_with_r2" \
--sjdbGTFfile $sjdbgtf \
--readFilesCommand 'gzip -c -d' \
--soloType CB_UMI_Simple \
--soloCBwhitelist $whitelists \
--soloUMIlen 12 \
--soloCBlen 16 \
--soloUMIstart 17 \
--soloCBmatchWLtype 1MM_multi_Nbase_pseudocounts \
--soloUMIfiltering MultiGeneUMI_CR \
--soloUMIdedup 1MM_CR \
--clipAdapterType CellRanger4 \
--outFilterScoreMin 30 \
--soloMultiMappers EM \
--soloCellFilter  EmptyDrops_CR \
--outBAMsortingThreadN 4 \
--outFileNamePrefix ${bam_filename%bam}


# COMMAND EXPLANATION
# ./bin/STAR \
# --runThreadN $threads \
# --runMode alignReads \
# --outSAMtype BAM Unsorted \
# --sysShell /bin/bash \
# --genomeDir "$star_reference_dir" \
# --readFilesIn "$files_with_r1" "$files_with_r2" \
# --sjdbGTFfile $sjdbgtf \
# --readFilesCommand 'gzip -c -d' \


# --soloType CB_UMI_Simple \
# --soloCBwhitelist $whitelists \
# --soloUMIlen 12 \
# --soloCBlen 16 \
# --soloUMIstart 17 \


# Matching CellRanger v4 and v5
# --soloCBmatchWLtype 1MM_multi_Nbase_pseudocounts \
# --soloUMIfiltering MultiGeneUMI_CR \
# --soloUMIdedup 1MM_CR \
# --clipAdapterType CellRanger4 \
# --outFilterScoreMin 30 \

# Distributes the multi-gene UMIs proportionally to the number of uniaue UMIs per gene
# --soloMultiMappers EM \

# Option for cell filtering nearly identical to that of CellRanger 3 and 4
# --soloCellFilter  EmptyDrops_CR \

# --outBAMsortingThreadN 4 \
# --outFileNamePrefix ${bam_filename%bam}