<?php

use Illuminate\Support\Facades\Route;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Facades\DB;

Route::get('/', function () {
    $cacheKey = 'laravel:visits';
    $visits = Redis::incr($cacheKey);

    return response()->json([
        'app' => config('app.name'),
        'env' => config('app.env'),
        'visits' => $visits,
        'cache' => 'redis',
        'db' => DB::connection()->getDatabaseName(),
    ]);
});

Route::get('/healthz', function () {
    $checks = [
        'app' => true,
        'redis' => false,
        'db' => false,
    ];

    try {
        Redis::ping();
        $checks['redis'] = true;
    } catch (\Throwable $e) {
        $checks['redis_error'] = $e->getMessage();
    }

    try {
        DB::connection()->getPdo();
        $checks['db'] = true;
    } catch (\Throwable $e) {
        $checks['db_error'] = $e->getMessage();
    }

    $healthy = $checks['redis'] && $checks['db'];

    return response()->json($checks, $healthy ? 200 : 503);
});
