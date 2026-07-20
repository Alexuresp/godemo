FROM composer:2 AS deps
WORKDIR /src
COPY composer.json composer.lock* ./
RUN composer install --no-dev --no-scripts --prefer-dist

FROM php:8.3-fpm-alpine
RUN apk add --no-cache \
    libpng-dev libjpeg-turbo-dev freetype-dev libwebp-dev libxpm-dev \
    libzip-dev oniguruma-dev libpq-dev libxml2-dev $PHPIZE_DEPS \
    && docker-php-ext-configure gd --with-jpeg --with-freetype --with-webp \
    && docker-php-ext-install -j$(nproc) gd mbstring bcmath pdo_pgsql zip opcache \
    && pecl install redis \
    && docker-php-ext-enable redis \
    && apk del $PHPIZE_DEPS

WORKDIR /var/www/html
COPY --from=deps /src/vendor ./vendor
COPY . .

RUN chown -R www-data:www-data storage bootstrap/cache

EXPOSE 9000
CMD ["php-fpm"]
