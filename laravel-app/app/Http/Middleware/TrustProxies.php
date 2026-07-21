<?php

namespace App\Http\Middleware;

use Illuminate\Http\Middleware\TrustProxies as Middleware;
use Symfony\Component\HttpFoundation\Request;

class TrustProxies extends Middleware
{
    protected $proxies = '10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,127.0.0.1/32';

    protected $headers = Request::HEADER_X_FORWARDED_FOR |
                        Request::HEADER_X_FORWARDED_PROTO |
                        Request::HEADER_X_FORWARDED_PORT;
}
