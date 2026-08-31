# 使用Node.js的官方镜像作为基础镜像
FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/node:22.19.0 AS node
# FROM node:latest AS node

# 设置工作目录为/app
WORKDIR /app

# 将当前目录下的所有文件复制到容器的/app目录中
COPY . .
# 安装项目依赖
# RUN npm install -g pnpm
RUN pnpm config set registry https://registry.npmmirror.com
WORKDIR /app/packages/ui
RUN pnpm install --frozen-lockfile
# RUN pnpm  build
RUN pnpm vite:build

FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/nginx:1.27.3
RUN rm /usr/local/nginx/conf/nginx.conf
COPY --from=node /app/packages/ui/nginx/nginx.conf /usr/local/nginx/conf/nginx.conf
COPY --from=node /app/packages/ui/distvite /usr/share/nginx/html

RUN chown -R 1000 /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html
EXPOSE 3000
CMD ["/usr/local/nginx/sbin/nginx", "-g", "daemon off;"]
RUN echo "deploy to nginx successful 🎉 🎉 🎉"