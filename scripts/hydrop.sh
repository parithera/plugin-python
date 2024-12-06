. ./config.sh

bam_filename="$outputdir/$name.bam";

star_reference_dir="indexes/$reference/index";
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

STAR \
--runThreadN $threads \
--runMode alignReads \
--outSAMtype BAM SortedByCoordinate \
--sysShell /bin/bash \
--genomeDir "$star_reference_dir" \
--readFilesIn "$files_with_r1" "$files_with_r2" \
--readFilesCommand 'gzip -c -d' \
--soloCBwhitelist $whitelists \
--soloType CB_UMI_Complex \
--soloCBposition 0_0_0_9 0_20_0_29 0_40_0_49 \
--soloUMIposition 0_50_0_57 \
--sjdbGTFfile $sjdbgtf \
--soloCellFilter CellRanger2.2 2000 0.99 10 \
--soloCBmatchWLtype 1MM \
--outFilterMultimapNmax 1 \
--outSAMattributes NH HI AS nM CB UB CR CY UR UY \
--outFileNamePrefix ${bam_filename%bam} \
--outReadsUnmapped Fastx \
--quantMode GeneCounts \
--bamRemoveDuplicatesType UniqueIdentical \
--soloFeatures Gene GeneFull Velocyto
