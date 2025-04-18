#ifndef DEEPX_TENSORFUNC_ELEMENTWISE_MIAOBYTE_BASIC_CU
#define DEEPX_TENSORFUNC_ELEMENTWISE_MIAOBYTE_BASIC_CU

#include <cuda_bf16.h>
#include <cuda_fp16.h>
#include "deepx/tensorfunc/cuda.hpp"
#include "deepx/tensorfunc/authors.hpp"

namespace deepx::tensorfunc
{
    // add
    template <typename T>
    __global__ void add_kernel(const T *A, const T *B, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] + B[idx];
        }
    }

    template <typename T>
    void launch_add(const T *a, const T *b, T *c, const int size)
    {
        // 启动kernel
        auto [numBlocks, blockSize] = BestDims(size);
        add_kernel<<<numBlocks, blockSize>>>(a, b, c, size);
        // 检查kernel执行是否成功
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch add kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }

    template void launch_add<double>(const double *a, const double *b, double *c, const int size);
    template void launch_add<float>(const float *a, const float *b, float *c, const int size);
    template void launch_add<half>(const half *a, const half *b, half *c, const int size);
    template void launch_add<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 *b, nv_bfloat16 *c, const int size);
    template void launch_add<int64_t>(const int64_t *a, const int64_t *b, int64_t *c, const int size);
    template void launch_add<int32_t>(const int32_t *a, const int32_t *b, int32_t *c, const int size);
    template void launch_add<int16_t>(const int16_t *a, const int16_t *b, int16_t *c, const int size);
    template void launch_add<int8_t>(const int8_t *a, const int8_t *b, int8_t *c, const int size);

    // addscalar
    template <typename T>
    __global__ void addscalar_kernel(const T *A, const T scalar, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] + scalar;
        }
    }

    template <typename T>
    void launch_addscalar(const T *a, const T scalar, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        addscalar_kernel<<<numBlocks, blockSize>>>(a, scalar, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch addscalar kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_addscalar<double>(const double *a, const double scalar, double *c, const int size);
    template void launch_addscalar<float>(const float *a, const float scalar, float *c, const int size);
    template void launch_addscalar<half>(const half *a, const half scalar, half *c, const int size);
    template void launch_addscalar<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 scalar, nv_bfloat16 *c, const int size);
    template void launch_addscalar<int64_t>(const int64_t *a, const int64_t scalar, int64_t *c, const int size);
    template void launch_addscalar<int32_t>(const int32_t *a, const int32_t scalar, int32_t *c, const int size);
    template void launch_addscalar<int16_t>(const int16_t *a, const int16_t scalar, int16_t *c, const int size);
    template void launch_addscalar<int8_t>(const int8_t *a, const int8_t scalar, int8_t *c, const int size);

    // sub
    template <typename T>
    __global__ void sub_kernel(const T *A, const T *B, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] - B[idx];
        }
    }

    template <typename T>
    void launch_sub(const T *a, const T *b, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        sub_kernel<<<numBlocks, blockSize>>>(a, b, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch sub kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_sub<double>(const double *a, const double *b, double *c, const int size);
    template void launch_sub<float>(const float *a, const float *b, float *c, const int size);
    template void launch_sub<half>(const half *a, const half *b, half *c, const int size);
    template void launch_sub<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 *b, nv_bfloat16 *c, const int size);
    template void launch_sub<int64_t>(const int64_t *a, const int64_t *b, int64_t *c, const int size);
    template void launch_sub<int32_t>(const int32_t *a, const int32_t *b, int32_t *c, const int size);
    template void launch_sub<int16_t>(const int16_t *a, const int16_t *b, int16_t *c, const int size);
    template void launch_sub<int8_t>(const int8_t *a, const int8_t *b, int8_t *c, const int size);

    // subscalar
    template <typename T>
    __global__ void subscalar_kernel(const T *A, const T scalar, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] - scalar;
        }
    }

    template <typename T>
    void launch_subscalar(const T *a, const T scalar, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        subscalar_kernel<<<numBlocks, blockSize>>>(a, scalar, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch subscalar kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_subscalar<double>(const double *a, const double scalar, double *c, const int size);
    template void launch_subscalar<float>(const float *a, const float scalar, float *c, const int size);
    template void launch_subscalar<half>(const half *a, const half scalar, half *c, const int size);
    template void launch_subscalar<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 scalar, nv_bfloat16 *c, const int size);
    template void launch_subscalar<int64_t>(const int64_t *a, const int64_t scalar, int64_t *c, const int size);
    template void launch_subscalar<int32_t>(const int32_t *a, const int32_t scalar, int32_t *c, const int size);
    template void launch_subscalar<int16_t>(const int16_t *a, const int16_t scalar, int16_t *c, const int size);
    template void launch_subscalar<int8_t>(const int8_t *a, const int8_t scalar, int8_t *c, const int size);

    // mul
    template <typename T>
    __global__ void mul_kernel(const T *A, const T *B, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] * B[idx];
        }
    }

    template <typename T>
    void launch_mul(const T *a, const T *b, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        mul_kernel<<<numBlocks, blockSize>>>(a, b, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch mul kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_mul<double>(const double *a, const double *b, double *c, const int size);
    template void launch_mul<float>(const float *a, const float *b, float *c, const int size);
    template void launch_mul<half>(const half *a, const half *b, half *c, const int size);
    template void launch_mul<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 *b, nv_bfloat16 *c, const int size);
    template void launch_mul<int64_t>(const int64_t *a, const int64_t *b, int64_t *c, const int size);
    template void launch_mul<int32_t>(const int32_t *a, const int32_t *b, int32_t *c, const int size);
    template void launch_mul<int16_t>(const int16_t *a, const int16_t *b, int16_t *c, const int size);
    template void launch_mul<int8_t>(const int8_t *a, const int8_t *b, int8_t *c, const int size);

    // mulscalar
    template <typename T>
    __global__ void mulscalar_kernel(const T *A, const T scalar, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] * scalar;
        }
    }

    template <typename T>
    void launch_mulscalar(const T *a, const T scalar, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        mulscalar_kernel<<<numBlocks, blockSize>>>(a, scalar, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch mulscalar kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_mulscalar<double>(const double *a, const double scalar, double *c, const int size);
    template void launch_mulscalar<float>(const float *a, const float scalar, float *c, const int size);
    template void launch_mulscalar<half>(const half *a, const half scalar, half *c, const int size);
    template void launch_mulscalar<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 scalar, nv_bfloat16 *c, const int size);
    template void launch_mulscalar<int64_t>(const int64_t *a, const int64_t scalar, int64_t *c, const int size);
    template void launch_mulscalar<int32_t>(const int32_t *a, const int32_t scalar, int32_t *c, const int size);
    template void launch_mulscalar<int16_t>(const int16_t *a, const int16_t scalar, int16_t *c, const int size);
    template void launch_mulscalar<int8_t>(const int8_t *a, const int8_t scalar, int8_t *c, const int size);

    // div
    template <typename T>
    __global__ void div_kernel(const T *A, const T *B, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] / B[idx];
        }
    }

    template <typename T>
    void launch_div(const T *a, const T *b, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        div_kernel<<<numBlocks, blockSize>>>(a, b, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch div kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_div<double>(const double *a, const double *b, double *c, const int size);
    template void launch_div<float>(const float *a, const float *b, float *c, const int size);
    template void launch_div<half>(const half *a, const half *b, half *c, const int size);
    template void launch_div<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 *b, nv_bfloat16 *c, const int size);
    template void launch_div<int64_t>(const int64_t *a, const int64_t *b, int64_t *c, const int size);
    template void launch_div<int32_t>(const int32_t *a, const int32_t *b, int32_t *c, const int size);
    template void launch_div<int16_t>(const int16_t *a, const int16_t *b, int16_t *c, const int size);
    template void launch_div<int8_t>(const int8_t *a, const int8_t *b, int8_t *c, const int size);

    // divscalar
    template <typename T>
    __global__ void divscalar_kernel(const T *A, const T scalar, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = A[idx] / scalar;
        }
    }

    template <typename T>
    void launch_divscalar(const T *a, const T scalar, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        divscalar_kernel<<<numBlocks, blockSize>>>(a, scalar, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch divscalar kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_divscalar<double>(const double *a, const double scalar, double *c, const int size);
    template void launch_divscalar<float>(const float *a, const float scalar, float *c, const int size);
    template void launch_divscalar<half>(const half *a, const half scalar, half *c, const int size);
    template void launch_divscalar<nv_bfloat16>(const nv_bfloat16 *a, const nv_bfloat16 scalar, nv_bfloat16 *c, const int size);
    template void launch_divscalar<int64_t>(const int64_t *a, const int64_t scalar, int64_t *c, const int size);
    template void launch_divscalar<int32_t>(const int32_t *a, const int32_t scalar, int32_t *c, const int size);
    template void launch_divscalar<int16_t>(const int16_t *a, const int16_t scalar, int16_t *c, const int size);
    template void launch_divscalar<int8_t>(const int8_t *a, const int8_t scalar, int8_t *c, const int size);

    // rdivscalar
    template <typename T>
    __global__ void rdivscalar_kernel(const T scalar, const T *A, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = scalar / A[idx];
        }
    }

    template <typename T>
    void launch_rdivscalar(const T scalar, const T *a, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        rdivscalar_kernel<<<numBlocks, blockSize>>>(scalar, a, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch rdivscalar kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_rdivscalar<double>(const double scalar, const double *a, double *c, const int size);
    template void launch_rdivscalar<float>(const float scalar, const float *a, float *c, const int size);
    template void launch_rdivscalar<half>(const half scalar, const half *a, half *c, const int size);
    template void launch_rdivscalar<nv_bfloat16>(const nv_bfloat16 scalar, const nv_bfloat16 *a, nv_bfloat16 *c, const int size);
    template void launch_rdivscalar<int64_t>(const int64_t scalar, const int64_t *a, int64_t *c, const int size);
    template void launch_rdivscalar<int32_t>(const int32_t scalar, const int32_t *a, int32_t *c, const int size);
    template void launch_rdivscalar<int16_t>(const int16_t scalar, const int16_t *a, int16_t *c, const int size);
    template void launch_rdivscalar<int8_t>(const int8_t scalar, const int8_t *a, int8_t *c, const int size);

    // invert
    template <typename T>
    __global__ void invert_kernel(const T *A, T *C, const int size)
    {
        for (int idx = blockIdx.x * blockDim.x + threadIdx.x; idx < size; idx += blockDim.x * gridDim.x)
        {
            C[idx] = ~A[idx];
        }
    }

    template <typename T>
    void launch_invert(const T *a, T *c, const int size)
    {
        auto [numBlocks, blockSize] = BestDims(size);
        invert_kernel<<<numBlocks, blockSize>>>(a, c, size);
        cudaError_t err = cudaGetLastError();
        if (err != cudaSuccess)
        {
            throw std::runtime_error("Failed to launch invert kernel: " +
                                     std::string(cudaGetErrorString(err)));
        }
    }
    template void launch_invert<int64_t>(const int64_t *a, int64_t *c, const int size);
    template void launch_invert<int32_t>(const int32_t *a, int32_t *c, const int size);
    template void launch_invert<int16_t>(const int16_t *a, int16_t *c, const int size);
    template void launch_invert<int8_t>(const int8_t *a, int8_t *c, const int size);

}

#endif // DEEPX_TENSORFUNC_ELEMENTWISE_MIAOBYTE_BASIC_CU
