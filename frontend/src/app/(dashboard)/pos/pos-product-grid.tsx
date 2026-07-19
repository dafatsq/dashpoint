"use client";

import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Loader2, Search, ShoppingCart } from "lucide-react";
import type { Product } from "@/types";

import { classifyStock, formatCurrency, getImageUrl, getProductPrice, getProductQuantity } from "./pos-helpers";
import type { ProductGridProps } from "./pos-types";

function ProductCard({
  product,
  disabled,
  onClick,
}: {
  product: Product;
  disabled: boolean;
  onClick: () => void;
}) {
  const { isLowStock, isOutOfStock } = classifyStock(product);
  const quantity = getProductQuantity(product);
  const price = getProductPrice(product);

  return (
    <button
      onClick={onClick}
      disabled={disabled || isOutOfStock}
      className={`bg-card border rounded-lg p-3 text-left hover:border-primary hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed relative overflow-hidden ${
        isOutOfStock ? "opacity-60 bg-muted/50" : ""
      }`}
    >
      {(isLowStock || isOutOfStock) ? (
        <div
          className={`absolute top-2 right-2 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded-full shadow-sm z-10 ${
            isOutOfStock ? "bg-red-600 text-white" : "bg-yellow-500 text-white"
          }`}
        >
          {isOutOfStock ? "Out of Stock" : "Low Stock"}
        </div>
      ) : null}

      <div className="aspect-square bg-muted rounded-md mb-2 flex items-center justify-center overflow-hidden">
        {product.image_url ? (
          <img
            src={getImageUrl(product.image_url)}
            alt={product.name}
            className={`w-full h-full object-cover ${isOutOfStock ? "grayscale" : ""}`}
            onError={(event) => {
              (event.target as HTMLImageElement).style.display = "none";
              ((event.target as HTMLImageElement).nextSibling as HTMLElement).style.display = "flex";
            }}
          />
        ) : null}
        <ShoppingCart
          className="h-8 w-8 text-muted-foreground"
          style={{ display: product.image_url ? "none" : "block" }}
        />
      </div>
      <p className="font-medium text-sm break-words">{product.name}</p>
      <p className="text-xs text-muted-foreground">{product.category_name}</p>
      <div className="flex flex-col gap-1 mt-2">
        <span className="font-bold text-primary font-mono">{formatCurrency(price)}</span>
        <span
          className={`text-xs font-medium ${
            isOutOfStock ? "text-red-600" : isLowStock ? "text-yellow-600" : "text-muted-foreground"
          }`}
        >
          Stock: {quantity}
        </span>
      </div>
    </button>
  );
}

export function PosProductGrid({
  products,
  categories,
  currentShift,
  isLoadingProducts,
  isFetchingMore,
  hasMore,
  searchQuery,
  selectedCategory,
  onSearchChange,
  onCategoryChange,
  onAddToCart,
  loadMoreRef,
}: ProductGridProps) {
  return (
    <div className="flex-1 flex flex-col p-4 overflow-hidden pb-16 lg:pb-4">
      <div className="flex gap-4 mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search products or scan barcode..."
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            className="pl-9"
          />
        </div>
        <Select value={selectedCategory} onValueChange={onCategoryChange}>
          <SelectTrigger className="w-48">
            <SelectValue placeholder="All Categories" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {categories.map((category) => (
              <SelectItem key={category.id} value={category.id}>
                {category.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoadingProducts ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : products.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
            <ShoppingCart className="h-12 w-12 mb-2" />
            <p>No products found</p>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3">
              {products.map((product) => (
                <ProductCard
                  key={product.id}
                  product={product}
                  disabled={!currentShift}
                  onClick={() => onAddToCart(product)}
                />
              ))}
            </div>

            <div ref={loadMoreRef} className="h-1 w-full" />
            <div className="py-4 flex items-center justify-center">
              {isFetchingMore ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading more products...
                </div>
              ) : !hasMore && products.length > 0 ? (
                <p className="text-xs text-muted-foreground">End of product list</p>
              ) : null}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
