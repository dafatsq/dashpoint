export const MAX_IMAGE_UPLOAD_SIZE = 5 * 1024 * 1024;

interface UploadImageBlobOptions {
  blob: Blob;
  fetchImpl?: typeof fetch;
  apiBaseUrl: string;
  accessToken: string | null;
}

export function validateImageFile(file: File): string | null {
  if (!file.type.startsWith("image/")) {
    return "Please select an image file";
  }

  if (file.size > MAX_IMAGE_UPLOAD_SIZE) {
    return "Image must be less than 5MB";
  }

  return null;
}

export async function uploadImageBlob({
  blob,
  fetchImpl = fetch,
  apiBaseUrl,
  accessToken,
}: UploadImageBlobOptions): Promise<string> {
  const formData = new FormData();
  formData.append("image", blob, "cropped-image.jpg");

  const response = await fetchImpl(`${apiBaseUrl}/upload/image`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken ?? ""}`,
    },
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || "Upload failed");
  }

  const data = await response.json();
  return data.url;
}

export function createImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.addEventListener("load", () => resolve(image));
    image.addEventListener("error", (error) => reject(error));
    image.src = url;
  });
}

export async function getCroppedImg(
  imageSrc: string,
  pixelCrop: { x: number; y: number; width: number; height: number },
): Promise<Blob> {
  const image = await createImage(imageSrc);
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  if (!ctx) {
    throw new Error("No 2d context");
  }

  canvas.width = pixelCrop.width;
  canvas.height = pixelCrop.height;

  ctx.drawImage(
    image,
    pixelCrop.x,
    pixelCrop.y,
    pixelCrop.width,
    pixelCrop.height,
    0,
    0,
    pixelCrop.width,
    pixelCrop.height,
  );

  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error("Canvas is empty"));
        return;
      }
      resolve(blob);
    }, "image/jpeg", 0.95);
  });
}
