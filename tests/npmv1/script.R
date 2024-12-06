
args <- commandArgs(trailingOnly = TRUE)
setwd(args[1])

library(Seurat)
library(hdf5r)
library(tidyverse)
library(devtools)
library(ggplot2)
library(rjson)

final_seurat_object <- readRDS("final_seurat_object.rds")

# Generate UMAP plot
umap_plot <- DimPlot(final_seurat_object, reduction = "umap")

# Save the plot as PNG
ggsave("graph.png", plot = umap_plot)
