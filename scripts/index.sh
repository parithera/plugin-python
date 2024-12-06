. ./config.sh

STAR \
--runThreadN $threads \
--runMode genomeGenerate \
--genomeDir $indexdir \
--genomeFastaFiles $fa \
--sjdbGTFfile $gtf \
--sjdbOverhang 100
