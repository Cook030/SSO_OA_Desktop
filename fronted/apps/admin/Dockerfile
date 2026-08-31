# ============================================
# 优化的多阶段 Dockerfile for pnpm monorepo
# ============================================

# -------------------- 阶段 1: 依赖安装 --------------------
    FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/node:22.19.0 AS deps

    WORKDIR /app
    
    # 复制 pnpm 配置和 workspace 定义（利用 Docker 缓存）
    COPY .npmrc pnpm-workspace.yaml package.json pnpm-lock.yaml ./
    
    # 复制所有 package.json 文件以利用 Docker 缓存
    # 只有在依赖变化时才会重新安装
    COPY packages/utils/package.json ./packages/utils/
    COPY packages/types/package.json ./packages/types/
    COPY packages/apis/package.json ./packages/apis/
    COPY packages/ui/package.json ./packages/ui/
    COPY apps/admin/package.json ./apps/admin/
    
    # 设置 pnpm 配置并安装所有依赖
    RUN pnpm config set registry https://registry.npmmirror.com
    RUN pnpm fetch
    RUN pnpm install --offline --frozen-lockfile
    
    # -------------------- 阶段 2: 构建应用 --------------------
    FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/node:22.19.0 AS builder
    
    WORKDIR /app
    
    # 从 deps 阶段复制 node_modules（利用缓存）
    # 说明：pnpm install 会在每个 workspace 包下生成独立的 node_modules（软链接到 .pnpm），
    # 必须从 deps 阶段一并拷贝过来，否则 packages/ui 使用 rollup 等依赖时会找不到。
    # 后续 COPY packages/* 会从 build context 叠加源码（context 中 node_modules 已被 .dockerignore 排除）。
    COPY --from=deps /app/node_modules ./node_modules
    COPY --from=deps /app/packages ./packages
    COPY --from=deps /app/apps/admin/node_modules ./apps/admin/node_modules
    
    # 复制必要的配置文件
    COPY .npmrc pnpm-workspace.yaml package.json pnpm-lock.yaml ./
    COPY tsconfig.base.json tsconfig.json ./
    COPY postcss.config.js tailwind.config.js ./
    
    # 复制 packages 源码（workspace 依赖）
    COPY packages/utils ./packages/utils
    COPY packages/types ./packages/types
    COPY packages/apis ./packages/apis
    COPY packages/ui ./packages/ui
    
    # 复制 admin 应用源码
    COPY apps/admin ./apps/admin
    
    # 构建 workspace 依赖（按依赖顺序）
    RUN cd packages/types && pnpm build && \
        cd ../utils && pnpm build && \
        cd ../apis && pnpm build && \
        cd ../ui && pnpm build
    
    # 构建 admin 应用
    WORKDIR /app/apps/admin
    RUN pnpm build
    
    # -------------------- 阶段 3: 生产镜像 --------------------
    FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/nginx:1.27.3 AS production
    
    # 删除默认 nginx 配置
    RUN rm -f /usr/local/nginx/conf/nginx.conf
    
    # 复制 nginx 配置
    COPY --from=builder /app/apps/admin/nginx/nginx.conf /usr/local/nginx/conf/nginx.conf
    
    # 复制构建产物
    COPY --from=builder /app/apps/admin/dist /usr/share/nginx/html/admin/
    
    # 设置正确的权限
    RUN chown -R 1000:1000 /usr/share/nginx/html && \
        chmod -R 755 /usr/share/nginx/html
    
    # 健康检查
    HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
        CMD curl -f http://localhost:8001/ || exit 1
    
    EXPOSE 8001
    
    CMD ["/usr/local/nginx/sbin/nginx", "-g", "daemon off;"]
    
    # 构建成功标记
    RUN echo "✅ admin application deployed to nginx successfully!"