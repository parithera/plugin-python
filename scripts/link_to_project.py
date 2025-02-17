import scanpy as sc
import anndata as ad
import sys
import hdf5plugin
import os
import json

if __name__=='__main__':
    output_path = sys.argv[1]
    sc.settings.figdir = output_path

    organization_folder = output_path.split("projects")[0]
    
    with open(os.path.join(output_path, "groups.json")) as f:
        samples = json.load(f)

        adatas = {}

        for sample in samples:
            sample_adata = ad.io.read_h5ad(os.path.join(organization_folder, "samples", sample["files"][0], 'scanpy', 'out.h5'))
            sample_adata.var_names_make_unique()
            adatas[sample["name"]] = sample_adata

    # Assuming the data is concatenated and stored in `adatas_combined`
    adatas_combined = sc.concat(adatas, label='sample')

    adatas_combined.obs_names_make_unique()
    print(adatas_combined.obs["sample"].value_counts())

    # mitochondrial genes, "MT-" for human, "Mt-" for mouse
    adatas_combined.var["mt"] = adatas_combined.var_names.str.startswith("MT-")
    # ribosomal genes
    adatas_combined.var["ribo"] = adatas_combined.var_names.str.startswith(("RPS", "RPL"))
    # hemoglobin genes
    adatas_combined.var["hb"] = adatas_combined.var_names.str.contains("^HB[^(P)]")

    sc.pp.calculate_qc_metrics(
        adatas_combined, qc_vars=["mt", "ribo", "hb"], inplace=True, log1p=True
    )

    sc.pp.filter_cells(adatas_combined, min_genes=100)
    sc.pp.filter_genes(adatas_combined, min_cells=3)

    sc.pp.scrublet(adatas_combined, batch_key="sample")

    # Saving count data
    adatas_combined.layers["counts"] = adatas_combined.X.copy()

    # Normalizing to median total counts
    sc.pp.normalize_total(adatas_combined)
    # Logarithmize the data
    sc.pp.log1p(adatas_combined)

    sc.pp.highly_variable_genes(adatas_combined, n_top_genes=2000, batch_key="sample")

    sc.tl.pca(adatas_combined)
    sc.pp.neighbors(adatas_combined)
    sc.tl.umap(adatas_combined)
    
    # Compute t-SNE
    sc.tl.tsne(adatas_combined)

    # Using the igraph implementation and a fixed number of iterations can be significantly faster, especially for larger datasets
    sc.tl.leiden(adatas_combined, flavor="igraph", n_iterations=2)

    for res in [0.02, 0.5, 2.0]:
        sc.tl.leiden(
            adatas_combined, key_added=f"leiden_res_{res:4.2f}", resolution=res, flavor="igraph"
        )

    # Obtain cluster-specific differentially expressed genes
    sc.tl.rank_genes_groups(adatas_combined, groupby="leiden_res_0.50", method="wilcoxon")

    adatas_combined.write_h5ad(
        output_path.replace("python", "out.h5ad"),
        compression=hdf5plugin.FILTERS["zstd"]
    )
