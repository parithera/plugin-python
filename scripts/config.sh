# With 10x r2 and r1 are switched
experiment="EX-810_S2"
inverted=true
reference="nih_GRCh38_human"
whitelists_folder="resources/10x-2016"
name="EX-810_S2"

# experiment="SRR14668399"
# inverted=false
# reference="nih_GRCm39_mouse"
# whitelists_folder="resources/hydrop"
# name="mouse"

# OUTPUT
outputdir="output/$name"
mkdir $outputdir

# DON'T TOUCH OR IT WILL EXPLODE
indexdir="indexes/${reference}/index"
fa="indexes/${reference}/file.fa"
gtf="indexes/${reference}/file.gtf"

threads=16