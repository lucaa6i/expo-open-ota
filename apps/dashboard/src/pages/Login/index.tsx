import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card.tsx';
import { Input } from '@/components/ui/input.tsx';
import { Button } from '@/components/ui/button.tsx';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Form, FormControl, FormField, FormItem, FormMessage } from '@/components/ui/form.tsx';
import { useCallback } from 'react';
import { setTokens } from '@/lib/auth.ts';
import { useNavigate } from 'react-router';
import { api } from '@/lib/api.ts';

const FormSchema = z.object({
  password: z.string().min(1, {
    message: 'Password is required',
  }),
});

export const Login = () => {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      password: '',
    },
  });
  const navigate = useNavigate();

  const onSubmit = useCallback(
    async (data: z.infer<typeof FormSchema>) => {
      try {
        const response = await api.login(data.password);
        setTokens(response.token, response.refreshToken);
        navigate('/');
      } catch {
        form.setError('password', {
          type: 'server',
          message: 'Error logging in',
        });
      }
    },
    [form]
  );

  return (
    <div className="flex-1 w-full h-screen flex items-center justify-center">
      <Card className="w-[350px]">
        <CardHeader>
          <CardTitle>Admin password</CardTitle>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="w-full gap-5 flex flex-col">
              <FormField
                control={form.control}
                name="password"
                render={({ field, fieldState }) => {
                  return (
                    <FormItem>
                      <FormControl>
                        <Input type={'password'} {...field} />
                      </FormControl>
                      <FormMessage>{fieldState.error?.message}</FormMessage>
                    </FormItem>
                  );
                }}
              />
              <Button type="submit">Submit</Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
};
