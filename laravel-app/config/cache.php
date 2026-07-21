<?php

use Illuminate\Support\Env;

return [

    'default' => env('CACHE_DRIVER', 'redis'),

    'stores' => [

        'redis' => [
            'driver' => 'redis',
            'connection' => 'default',
        ],

    ],

    'prefix' => env('CACHE_PREFIX', 'laravel_cache'),

];
