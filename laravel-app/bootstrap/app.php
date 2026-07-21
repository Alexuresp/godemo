<?php

use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;
use Illuminate\View\ViewServiceProvider;

$app = Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        api: __DIR__.'/../routes/api.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware) {
        $middleware->trustProxies('10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,127.0.0.1/32');
    })
    ->withExceptions(function (Exceptions $exceptions) {
        // Placeholder for exception handler configuration
    })
    ->create();

$app->register(ViewServiceProvider::class);

return $app;
