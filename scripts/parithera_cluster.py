
import scanpy as sc
import anndata as ad
import sys
import hdf5plugin
import os
import json

if __name__=='__main__':
    output_path = sys.argv[1]
    sc.settings.figdir = output_path

    sc.settings.set_figure_params(dpi=50, facecolor="white")
    adata = ad.read_h5ad(output_path.replace("python", "out.h5ad"))
    
    # Plot clusters
    sc.pl.umap(adata, color=["leiden"])

